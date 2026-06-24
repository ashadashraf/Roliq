package background

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roliq/roliq/internal/storage"
)

type Queue interface {
	SendMessage(context.Context, *sqs.SendMessageInput, ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}
type Worker struct {
	db          *pgxpool.Pool
	objects     storage.Store
	queue       Queue
	queueURL    string
	clamAddress string
	logger      *slog.Logger
}

func New(db *pgxpool.Pool, objects storage.Store, queue Queue, queueURL, clamAddress string, logger *slog.Logger) *Worker {
	return &Worker{db: db, objects: objects, queue: queue, queueURL: queueURL, clamAddress: clamAddress, logger: logger}
}

type scanJob struct{ OrganizationID, ResumeID, FileObjectID, ObjectKey, ContentType, ChecksumSHA256 string }

func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	w.process(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}
func (w *Worker) process(ctx context.Context) {
	if err := w.expireUploads(ctx); err != nil {
		w.logger.Error("upload_expiration_failed", "error", err)
	}
	if err := w.scanOne(ctx); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		w.logger.Error("resume_scan_cycle_failed", "error", err)
	}
	if w.queue != nil && w.queueURL != "" {
		if err := w.publishOne(ctx); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			w.logger.Error("outbox_publish_failed", "error", err)
		}
	}
}

func (w *Worker) scanOne(ctx context.Context) error {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var job scanJob
	err = tx.QueryRow(ctx, `SELECT r.organization_id,r.id,fo.id,fo.object_key,fo.content_type,fo.checksum_sha256 FROM resumes r JOIN resume_versions rv ON rv.resume_id=r.id AND rv.organization_id=r.organization_id AND rv.version_number=1 JOIN file_objects fo ON fo.id=rv.file_object_id AND fo.organization_id=rv.organization_id WHERE r.status='uploaded' AND fo.scan_status='pending' ORDER BY r.created_at FOR UPDATE OF r,fo SKIP LOCKED LIMIT 1`).Scan(&job.OrganizationID, &job.ResumeID, &job.FileObjectID, &job.ObjectKey, &job.ContentType, &job.ChecksumSHA256)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE resumes SET status='scanning',updated_at=now() WHERE id=$1`, job.ResumeID); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	body, err := w.objects.Download(ctx, job.ObjectKey)
	if err != nil {
		return w.finishScan(ctx, job, "failed", "storage download failed", err)
	}
	defer body.Close()
	data, err := io.ReadAll(io.LimitReader(body, 10*1024*1024+1))
	if err != nil {
		return w.finishScan(ctx, job, "failed", "file read failed", err)
	}
	if len(data) > 10*1024*1024 {
		return w.finishScan(ctx, job, "invalid", "file exceeds size limit", nil)
	}
	actualChecksum := fmt.Sprintf("%x", sha256.Sum256(data))
	if !strings.EqualFold(actualChecksum, job.ChecksumSHA256) {
		return w.finishScan(ctx, job, "invalid", "file checksum does not match the upload declaration", nil)
	}
	if err := validateDocument(data, job.ContentType); err != nil {
		return w.finishScan(ctx, job, "invalid", err.Error(), nil)
	}
	clean, detail, err := scanClamAV(ctx, w.clamAddress, data)
	if err != nil {
		return w.finishScan(ctx, job, "failed", "malware scanner unavailable", err)
	}
	if !clean {
		return w.finishScan(ctx, job, "infected", detail, nil)
	}
	return w.finishScan(ctx, job, "clean", "", nil)
}

func (w *Worker) expireUploads(ctx context.Context) error {
	_, err := w.db.Exec(ctx, `WITH expired AS (
		UPDATE resume_uploads SET completed_at=now()
		WHERE completed_at IS NULL AND expires_at<now()
		RETURNING resume_id, file_object_id
	) UPDATE resumes r SET status='failed',rejection_reason='Upload session expired before completion',updated_at=now()
	FROM expired e WHERE r.id=e.resume_id AND r.status='pending'`)
	return err
}

func validateDocument(data []byte, contentType string) error {
	switch contentType {
	case "application/pdf":
		if len(data) < 5 || !bytes.Equal(data[:5], []byte("%PDF-")) {
			return fmt.Errorf("file signature is not a PDF")
		}
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return fmt.Errorf("file is not a valid DOCX archive")
		}
		found := false
		for _, file := range reader.File {
			if file.Name == "word/document.xml" {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("DOCX document content is missing")
		}
	default:
		return fmt.Errorf("unsupported content type")
	}
	return nil
}

func scanClamAV(ctx context.Context, address string, data []byte) (bool, string, error) {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return false, "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	if _, err = conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return false, "", err
	}
	for offset := 0; offset < len(data); {
		end := offset + 32*1024
		if end > len(data) {
			end = len(data)
		}
		size := make([]byte, 4)
		binary.BigEndian.PutUint32(size, uint32(end-offset))
		if _, err = conn.Write(size); err != nil {
			return false, "", err
		}
		if _, err = conn.Write(data[offset:end]); err != nil {
			return false, "", err
		}
		offset = end
	}
	if _, err = conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return false, "", err
	}
	response, err := io.ReadAll(io.LimitReader(conn, 4096))
	if err != nil {
		return false, "", err
	}
	detail := normalizeClamAVDetail(response)
	if strings.Contains(detail, "FOUND") {
		return false, detail, nil
	}
	if strings.Contains(detail, "OK") {
		return true, detail, nil
	}
	return false, detail, fmt.Errorf("unexpected ClamAV response: %s", detail)
}

func normalizeClamAVDetail(response []byte) string {
	return strings.TrimSpace(strings.ReplaceAll(string(response), "\x00", ""))
}

func (w *Worker) finishScan(ctx context.Context, job scanJob, status, detail string, cause error) error {
	resumeStatus := "ready"
	reason := any(nil)
	eventType := "resume.scan.completed.v1"
	if status != "clean" {
		resumeStatus = "rejected"
		reason = detail
		eventType = "resume.scan.rejected.v1"
	}
	if status == "failed" {
		resumeStatus = "failed"
	}
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE file_objects SET scan_status=$2::file_scan_status,scan_detail=NULLIF($3,''),scanned_at=now() WHERE id=$1`, job.FileObjectID, status, detail); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE resumes SET status=$2::resume_status,rejection_reason=$3,updated_at=now() WHERE id=$1`, job.ResumeID, resumeStatus, reason); err != nil {
		return err
	}
	payload := fmt.Sprintf(`{"organizationId":%q,"resumeId":%q,"fileObjectId":%q,"status":%q}`, job.OrganizationID, job.ResumeID, job.FileObjectID, status)
	if _, err = tx.Exec(ctx, `INSERT INTO outbox_events(id,organization_id,event_type,aggregate_type,aggregate_id,payload) VALUES($1,$2,$3,'resume',$4,$5::jsonb)`, uuid.Must(uuid.NewV7()), job.OrganizationID, eventType, job.ResumeID, payload); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	if cause != nil {
		return cause
	}
	w.logger.Info("resume_scan_finished", "resume_id", job.ResumeID, "status", status)
	return nil
}

func (w *Worker) publishOne(ctx context.Context) error {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var id, eventType string
	var payload []byte
	err = tx.QueryRow(ctx, `SELECT id,event_type,payload FROM outbox_events WHERE published_at IS NULL AND attempts<10 ORDER BY occurred_at FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&id, &eventType, &payload)
	if err != nil {
		return err
	}
	_, sendErr := w.queue.SendMessage(ctx, &sqs.SendMessageInput{QueueUrl: aws.String(w.queueURL), MessageBody: aws.String(string(payload)), MessageAttributes: map[string]types.MessageAttributeValue{"event_type": {DataType: aws.String("String"), StringValue: aws.String(eventType)}}})
	if sendErr != nil {
		_, _ = tx.Exec(ctx, `UPDATE outbox_events SET attempts=attempts+1,last_error=$2 WHERE id=$1`, id, truncate(sendErr.Error(), 1000))
		_ = tx.Commit(ctx)
		return sendErr
	}
	if _, err = tx.Exec(ctx, `UPDATE outbox_events SET published_at=now(),attempts=attempts+1,last_error=NULL WHERE id=$1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
