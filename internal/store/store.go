package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roliq/roliq/internal/auth"
	"github.com/roliq/roliq/internal/model"
)

var ErrNotFound = errors.New("not found")
var ErrIdempotencyConflict = errors.New("idempotency key was used for a different request")

type Store struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Store { return &Store{db: db} }

func (s *Store) Ping(ctx context.Context) error { return s.db.Ping(ctx) }

func newID() string { return uuid.Must(uuid.NewV7()).String() }

func setContext(ctx context.Context, tx pgx.Tx, userID, organizationID string) error {
	if _, err := tx.Exec(ctx, "SELECT set_config('app.user_id', $1, true)", userID); err != nil {
		return err
	}
	if organizationID != "" {
		if _, err := tx.Exec(ctx, "SELECT set_config('app.organization_id', $1, true)", organizationID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Bootstrap(ctx context.Context, claims auth.Claims, requestID string) (model.Session, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return model.Session{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, claims.Issuer+"\x00"+claims.Subject); err != nil {
		return model.Session{}, err
	}

	var user model.User
	err = tx.QueryRow(ctx, `SELECT u.id, u.email, u.display_name FROM auth_identities ai JOIN users u ON u.id=ai.user_id WHERE ai.issuer=$1 AND ai.subject=$2 AND u.deleted_at IS NULL`, claims.Issuer, claims.Subject).Scan(&user.ID, &user.Email, &user.DisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		user = model.User{ID: newID(), Email: claims.Email, DisplayName: strings.TrimSpace(claims.Name)}
		if user.DisplayName == "" {
			user.DisplayName = strings.Split(claims.Email, "@")[0]
		}
		orgID := newID()
		if err := setContext(ctx, tx, user.ID, orgID); err != nil {
			return model.Session{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO users(id,email,display_name) VALUES($1,$2,$3)`, user.ID, user.Email, user.DisplayName); err != nil {
			return model.Session{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO auth_identities(id,user_id,issuer,subject) VALUES($1,$2,$3,$4)`, newID(), user.ID, claims.Issuer, claims.Subject); err != nil {
			return model.Session{}, err
		}
		org := model.Organization{ID: orgID, Name: workspaceName(user.DisplayName), Slug: personalSlug(user.DisplayName, orgID), Type: "personal"}
		statements := []struct {
			sql  string
			args []any
		}{
			{`INSERT INTO organizations(id,name,slug,type) VALUES($1,$2,$3,'personal')`, []any{org.ID, org.Name, org.Slug}},
			{`INSERT INTO organization_memberships(id,organization_id,user_id,role) VALUES($1,$2,$3,'owner')`, []any{newID(), org.ID, user.ID}},
			{`INSERT INTO organization_subscriptions(id,organization_id,plan_code,status) VALUES($1,$2,'free','active')`, []any{newID(), org.ID}},
			{`INSERT INTO onboarding_progress(id,organization_id,user_id) VALUES($1,$2,$3)`, []any{newID(), org.ID, user.ID}},
			{`INSERT INTO career_profiles(id,organization_id,user_id) VALUES($1,$2,$3)`, []any{newID(), org.ID, user.ID}},
		}
		for _, statement := range statements {
			if _, err := tx.Exec(ctx, statement.sql, statement.args...); err != nil {
				return model.Session{}, err
			}
		}
		payload, _ := json.Marshal(map[string]any{"userId": user.ID, "organizationId": org.ID})
		if _, err := tx.Exec(ctx, `INSERT INTO outbox_events(id,organization_id,event_type,aggregate_type,aggregate_id,payload) VALUES($1,$2,'identity.user_bootstrapped.v1','user',$3,$4)`, newID(), org.ID, user.ID, payload); err != nil {
			return model.Session{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO audit_logs(id,organization_id,actor_user_id,action,resource_type,resource_id,request_id) VALUES($1,$2,$3,'identity.user_bootstrapped','user',$3,$4)`, newID(), org.ID, user.ID, requestID); err != nil {
			return model.Session{}, err
		}
		session := model.Session{User: user, Organization: org, Onboarding: model.Onboarding{CurrentStep: 1, Status: "not_started"}}
		if err := tx.Commit(ctx); err != nil {
			return model.Session{}, err
		}
		return session, nil
	}
	if err != nil {
		return model.Session{}, err
	}
	if err := setContext(ctx, tx, user.ID, ""); err != nil {
		return model.Session{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET email=$2, display_name=CASE WHEN $3='' THEN display_name ELSE $3 END, updated_at=now() WHERE id=$1`, user.ID, claims.Email, strings.TrimSpace(claims.Name)); err != nil {
		return model.Session{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE auth_identities SET last_seen_at=now() WHERE issuer=$1 AND subject=$2`, claims.Issuer, claims.Subject); err != nil {
		return model.Session{}, err
	}
	session, err := resolveInTx(ctx, tx, user, "")
	if err != nil {
		return model.Session{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Session{}, err
	}
	return session, nil
}

func (s *Store) ResolveSession(ctx context.Context, claims auth.Claims, requestedOrg string) (model.Session, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return model.Session{}, err
	}
	defer tx.Rollback(ctx)
	var user model.User
	if err := tx.QueryRow(ctx, `SELECT u.id,u.email,u.display_name FROM auth_identities ai JOIN users u ON u.id=ai.user_id WHERE ai.issuer=$1 AND ai.subject=$2 AND u.deleted_at IS NULL`, claims.Issuer, claims.Subject).Scan(&user.ID, &user.Email, &user.DisplayName); errors.Is(err, pgx.ErrNoRows) {
		return model.Session{}, ErrNotFound
	} else if err != nil {
		return model.Session{}, err
	}
	if err := setContext(ctx, tx, user.ID, ""); err != nil {
		return model.Session{}, err
	}
	session, err := resolveInTx(ctx, tx, user, requestedOrg)
	if err != nil {
		return model.Session{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Session{}, err
	}
	return session, nil
}

func resolveInTx(ctx context.Context, tx pgx.Tx, user model.User, requestedOrg string) (model.Session, error) {
	var orgID string
	if requestedOrg == "" {
		if err := tx.QueryRow(ctx, `SELECT organization_id FROM organization_memberships WHERE user_id=$1 ORDER BY created_at LIMIT 1`, user.ID).Scan(&orgID); errors.Is(err, pgx.ErrNoRows) {
			return model.Session{}, ErrNotFound
		} else if err != nil {
			return model.Session{}, err
		}
	} else {
		if _, err := uuid.Parse(requestedOrg); err != nil {
			return model.Session{}, ErrNotFound
		}
		if err := tx.QueryRow(ctx, `SELECT organization_id FROM organization_memberships WHERE user_id=$1 AND organization_id=$2`, user.ID, requestedOrg).Scan(&orgID); errors.Is(err, pgx.ErrNoRows) {
			return model.Session{}, ErrNotFound
		} else if err != nil {
			return model.Session{}, err
		}
	}
	if err := setContext(ctx, tx, user.ID, orgID); err != nil {
		return model.Session{}, err
	}
	var org model.Organization
	if err := tx.QueryRow(ctx, `SELECT id,name,slug,type::text FROM organizations WHERE id=$1 AND deleted_at IS NULL`, orgID).Scan(&org.ID, &org.Name, &org.Slug, &org.Type); err != nil {
		return model.Session{}, err
	}
	var onboarding model.Onboarding
	if err := tx.QueryRow(ctx, `SELECT current_step,status::text,profile_method::text,completed_at FROM onboarding_progress WHERE organization_id=$1 AND user_id=$2`, orgID, user.ID).Scan(&onboarding.CurrentStep, &onboarding.Status, &onboarding.ProfileMethod, &onboarding.CompletedAt); err != nil {
		return model.Session{}, err
	}
	return model.Session{User: user, Organization: org, Onboarding: onboarding}, nil
}

func personalSlug(name, id string) string {
	value := strings.ToLower(strings.TrimSpace(name))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if len(value) > 42 {
		value = value[:42]
	}
	if len(value) < 2 {
		value = "personal"
	}
	return value + "-" + strings.ReplaceAll(id, "-", "")[:10]
}

func workspaceName(displayName string) string {
	runes := []rune(strings.TrimSpace(displayName))
	if len(runes) > 105 {
		runes = runes[:105]
	}
	if len(runes) == 0 {
		return "Personal workspace"
	}
	return string(runes) + "'s workspace"
}

func (s *Store) tenantTx(ctx context.Context, session model.Session) (pgx.Tx, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	if err := setContext(ctx, tx, session.User.ID, session.Organization.ID); err != nil {
		tx.Rollback(ctx)
		return nil, err
	}
	return tx, nil
}

func (s *Store) GetProfile(ctx context.Context, session model.Session) (model.CareerProfile, error) {
	tx, err := s.tenantTx(ctx, session)
	if err != nil {
		return model.CareerProfile{}, err
	}
	defer tx.Rollback(ctx)
	profile, err := getProfileInTx(ctx, tx, session)
	if err != nil {
		return model.CareerProfile{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.CareerProfile{}, err
	}
	return profile, nil
}

func getProfileInTx(ctx context.Context, tx pgx.Tx, session model.Session) (model.CareerProfile, error) {
	var profile model.CareerProfile
	var profileID string
	if err := tx.QueryRow(ctx, `SELECT id,headline,summary,country_code,time_zone,city,years_experience,updated_at FROM career_profiles WHERE organization_id=$1 AND user_id=$2 AND deleted_at IS NULL`, session.Organization.ID, session.User.ID).Scan(&profileID, &profile.Headline, &profile.Summary, &profile.CountryCode, &profile.TimeZone, &profile.City, &profile.YearsExperience, &profile.UpdatedAt); err != nil {
		return profile, err
	}
	profile.Skills = []string{}
	rows, err := tx.Query(ctx, `SELECT name FROM profile_skills WHERE organization_id=$1 AND career_profile_id=$2 ORDER BY position,id`, session.Organization.ID, profileID)
	if err != nil {
		return profile, err
	}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			rows.Close()
			return profile, err
		}
		profile.Skills = append(profile.Skills, value)
	}
	rows.Close()
	profile.Experiences = []model.Experience{}
	rows, err = tx.Query(ctx, `SELECT id,company,title,location,start_date::text,end_date::text,is_current,description FROM work_experiences WHERE organization_id=$1 AND career_profile_id=$2 ORDER BY position,id`, session.Organization.ID, profileID)
	if err != nil {
		return profile, err
	}
	for rows.Next() {
		var item model.Experience
		if err := rows.Scan(&item.ID, &item.Company, &item.Title, &item.Location, &item.StartDate, &item.EndDate, &item.IsCurrent, &item.Description); err != nil {
			rows.Close()
			return profile, err
		}
		profile.Experiences = append(profile.Experiences, item)
	}
	rows.Close()
	profile.Education = []model.Education{}
	rows, err = tx.Query(ctx, `SELECT id,institution,degree,field_of_study,start_date::text,end_date::text FROM educations WHERE organization_id=$1 AND career_profile_id=$2 ORDER BY position,id`, session.Organization.ID, profileID)
	if err != nil {
		return profile, err
	}
	for rows.Next() {
		var item model.Education
		if err := rows.Scan(&item.ID, &item.Institution, &item.Degree, &item.FieldOfStudy, &item.StartDate, &item.EndDate); err != nil {
			rows.Close()
			return profile, err
		}
		profile.Education = append(profile.Education, item)
	}
	rows.Close()
	return profile, rows.Err()
}

func (s *Store) SaveProfile(ctx context.Context, session model.Session, profile model.CareerProfile, requestID string) (model.CareerProfile, error) {
	tx, err := s.tenantTx(ctx, session)
	if err != nil {
		return model.CareerProfile{}, err
	}
	defer tx.Rollback(ctx)
	var profileID string
	if err := tx.QueryRow(ctx, `UPDATE career_profiles SET headline=$3,summary=$4,country_code=$5,time_zone=$6,city=$7,years_experience=$8,version=version+1,updated_at=now() WHERE organization_id=$1 AND user_id=$2 AND deleted_at IS NULL RETURNING id`, session.Organization.ID, session.User.ID, profile.Headline, profile.Summary, profile.CountryCode, profile.TimeZone, profile.City, profile.YearsExperience).Scan(&profileID); err != nil {
		return model.CareerProfile{}, err
	}
	for _, table := range []string{"profile_skills", "work_experiences", "educations"} {
		if _, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE organization_id=$1 AND career_profile_id=$2", session.Organization.ID, profileID); err != nil {
			return model.CareerProfile{}, err
		}
	}
	seen := map[string]bool{}
	for position, skill := range profile.Skills {
		name := strings.TrimSpace(skill)
		normalized := strings.ToLower(name)
		if name == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		if _, err := tx.Exec(ctx, `INSERT INTO profile_skills(id,organization_id,career_profile_id,name,normalized_name,position) VALUES($1,$2,$3,$4,$5,$6)`, newID(), session.Organization.ID, profileID, name, normalized, position); err != nil {
			return model.CareerProfile{}, err
		}
	}
	for position, item := range profile.Experiences {
		id := item.ID
		if _, err := uuid.Parse(id); err != nil {
			id = newID()
		}
		if _, err := tx.Exec(ctx, `INSERT INTO work_experiences(id,organization_id,career_profile_id,company,title,location,start_date,end_date,is_current,description,position) VALUES($1,$2,$3,$4,$5,$6,$7::date,$8::date,$9,$10,$11)`, id, session.Organization.ID, profileID, item.Company, item.Title, item.Location, item.StartDate, item.EndDate, item.IsCurrent, item.Description, position); err != nil {
			return model.CareerProfile{}, err
		}
	}
	for position, item := range profile.Education {
		id := item.ID
		if _, err := uuid.Parse(id); err != nil {
			id = newID()
		}
		if _, err := tx.Exec(ctx, `INSERT INTO educations(id,organization_id,career_profile_id,institution,degree,field_of_study,start_date,end_date,position) VALUES($1,$2,$3,$4,$5,$6,$7::date,$8::date,$9)`, id, session.Organization.ID, profileID, item.Institution, item.Degree, item.FieldOfStudy, item.StartDate, item.EndDate, position); err != nil {
			return model.CareerProfile{}, err
		}
	}
	payload, _ := json.Marshal(map[string]any{"profileId": profileID, "userId": session.User.ID})
	if _, err := tx.Exec(ctx, `INSERT INTO outbox_events(id,organization_id,event_type,aggregate_type,aggregate_id,payload) VALUES($1,$2,'profile.updated.v1','career_profile',$3,$4)`, newID(), session.Organization.ID, profileID, payload); err != nil {
		return model.CareerProfile{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO audit_logs(id,organization_id,actor_user_id,action,resource_type,resource_id,request_id) VALUES($1,$2,$3,'profile.updated','career_profile',$4,$5)`, newID(), session.Organization.ID, session.User.ID, profileID, requestID); err != nil {
		return model.CareerProfile{}, err
	}
	result, err := getProfileInTx(ctx, tx, session)
	if err != nil {
		return model.CareerProfile{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.CareerProfile{}, err
	}
	return result, nil
}

func (s *Store) UpdateOnboarding(ctx context.Context, session model.Session, step *int, status, method *string, requestID string) (model.Onboarding, error) {
	tx, err := s.tenantTx(ctx, session)
	if err != nil {
		return model.Onboarding{}, err
	}
	defer tx.Rollback(ctx)
	var result model.Onboarding
	err = tx.QueryRow(ctx, `UPDATE onboarding_progress SET current_step=COALESCE($3,current_step),status=COALESCE($4::onboarding_status,status),profile_method=COALESCE($5::profile_method,profile_method),completed_at=CASE WHEN $4='completed' THEN COALESCE(completed_at,now()) ELSE completed_at END,updated_at=now() WHERE organization_id=$1 AND user_id=$2 RETURNING current_step,status::text,profile_method::text,completed_at`, session.Organization.ID, session.User.ID, step, status, method).Scan(&result.CurrentStep, &result.Status, &result.ProfileMethod, &result.CompletedAt)
	if err != nil {
		return model.Onboarding{}, err
	}
	if result.Status == "completed" && session.Onboarding.Status != "completed" {
		payload, _ := json.Marshal(map[string]any{"userId": session.User.ID})
		if _, err := tx.Exec(ctx, `INSERT INTO outbox_events(id,organization_id,event_type,aggregate_type,aggregate_id,payload) VALUES($1,$2,'onboarding.completed.v1','user',$3,$4)`, newID(), session.Organization.ID, session.User.ID, payload); err != nil {
			return model.Onboarding{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO audit_logs(id,organization_id,actor_user_id,action,resource_type,resource_id,request_id) VALUES($1,$2,$3,'onboarding.completed','user',$3,$4)`, newID(), session.Organization.ID, session.User.ID, requestID); err != nil {
			return model.Onboarding{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Onboarding{}, err
	}
	return result, nil
}

type Upload struct {
	ID, OrganizationID, ResumeID, FileObjectID, ObjectKey, FileName, ContentType, Checksum string
	SizeBytes                                                                              int64
	ExpiresAt                                                                              time.Time
}

func (s *Store) CreateUpload(ctx context.Context, session model.Session, fileName, contentType, checksum, bucket, objectKey string, size int64, expires time.Time, requestID, idempotencyKey, requestHash string) (Upload, error) {
	tx, err := s.tenantTx(ctx, session)
	if err != nil {
		return Upload{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, session.Organization.ID+"\x00"+idempotencyKey); err != nil {
		return Upload{}, err
	}
	var existingHash string
	var existingBody []byte
	err = tx.QueryRow(ctx, `SELECT request_hash,response_body FROM idempotency_keys WHERE organization_id=$1 AND key=$2 AND expires_at>now()`, session.Organization.ID, idempotencyKey).Scan(&existingHash, &existingBody)
	if err == nil {
		if existingHash != requestHash {
			return Upload{}, ErrIdempotencyConflict
		}
		var existing Upload
		if err := json.Unmarshal(existingBody, &existing); err != nil {
			return Upload{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return Upload{}, err
		}
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Upload{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM idempotency_keys WHERE organization_id=$1 AND key=$2 AND expires_at<=now()`, session.Organization.ID, idempotencyKey); err != nil {
		return Upload{}, err
	}
	upload := Upload{ID: newID(), OrganizationID: session.Organization.ID, ResumeID: newID(), FileObjectID: newID(), ObjectKey: objectKey, FileName: fileName, ContentType: contentType, Checksum: checksum, SizeBytes: size, ExpiresAt: expires}
	if _, err = tx.Exec(ctx, `INSERT INTO file_objects(id,organization_id,bucket,object_key,original_name,content_type,size_bytes,checksum_sha256) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, upload.FileObjectID, session.Organization.ID, bucket, objectKey, fileName, contentType, size, checksum); err != nil {
		return Upload{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO resumes(id,organization_id,user_id,title) VALUES($1,$2,$3,$4)`, upload.ResumeID, session.Organization.ID, session.User.ID, fileName); err != nil {
		return Upload{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO resume_versions(id,organization_id,resume_id,file_object_id,source,version_number,created_by) VALUES($1,$2,$3,$4,'original',1,$5)`, newID(), session.Organization.ID, upload.ResumeID, upload.FileObjectID, session.User.ID); err != nil {
		return Upload{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO resume_uploads(id,organization_id,resume_id,file_object_id,expires_at) VALUES($1,$2,$3,$4,$5)`, upload.ID, session.Organization.ID, upload.ResumeID, upload.FileObjectID, expires); err != nil {
		return Upload{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(id,organization_id,actor_user_id,action,resource_type,resource_id,request_id,metadata) VALUES($1,$2,$3,'resume.upload_created','resume',$4,$5,jsonb_build_object('contentType',$6,'sizeBytes',$7))`, newID(), session.Organization.ID, session.User.ID, upload.ResumeID, requestID, contentType, size); err != nil {
		return Upload{}, err
	}
	responseBody, _ := json.Marshal(upload)
	if _, err = tx.Exec(ctx, `INSERT INTO idempotency_keys(organization_id,key,request_hash,response_status,response_body,expires_at) VALUES($1,$2,$3,201,$4,$5)`, session.Organization.ID, idempotencyKey, requestHash, responseBody, expires); err != nil {
		return Upload{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Upload{}, err
	}
	return upload, nil
}

func (s *Store) GetUpload(ctx context.Context, session model.Session, uploadID string) (Upload, error) {
	tx, err := s.tenantTx(ctx, session)
	if err != nil {
		return Upload{}, err
	}
	defer tx.Rollback(ctx)
	var u Upload
	err = tx.QueryRow(ctx, `SELECT ru.id,ru.organization_id,ru.resume_id,ru.file_object_id,fo.object_key,fo.original_name,fo.content_type,fo.checksum_sha256,fo.size_bytes,ru.expires_at FROM resume_uploads ru JOIN file_objects fo ON fo.id=ru.file_object_id AND fo.organization_id=ru.organization_id WHERE ru.organization_id=$1 AND ru.id=$2 AND ru.completed_at IS NULL`, session.Organization.ID, uploadID).Scan(&u.ID, &u.OrganizationID, &u.ResumeID, &u.FileObjectID, &u.ObjectKey, &u.FileName, &u.ContentType, &u.Checksum, &u.SizeBytes, &u.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Upload{}, ErrNotFound
	}
	if err != nil {
		return Upload{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Upload{}, err
	}
	return u, nil
}

func (s *Store) CompleteUpload(ctx context.Context, session model.Session, u Upload, requestID string) (model.Resume, error) {
	tx, err := s.tenantTx(ctx, session)
	if err != nil {
		return model.Resume{}, err
	}
	defer tx.Rollback(ctx)
	if u.ExpiresAt.Before(time.Now()) {
		return model.Resume{}, fmt.Errorf("upload expired")
	}
	var completed *time.Time
	if err = tx.QueryRow(ctx, `SELECT completed_at FROM resume_uploads WHERE organization_id=$1 AND id=$2 FOR UPDATE`, session.Organization.ID, u.ID).Scan(&completed); err != nil {
		return model.Resume{}, err
	}
	if completed != nil {
		return model.Resume{}, fmt.Errorf("upload already completed")
	}
	if _, err = tx.Exec(ctx, `UPDATE resume_uploads SET completed_at=now() WHERE organization_id=$1 AND id=$2`, session.Organization.ID, u.ID); err != nil {
		return model.Resume{}, err
	}
	var result model.Resume
	if err = tx.QueryRow(ctx, `UPDATE resumes SET status='uploaded',updated_at=now() WHERE organization_id=$1 AND id=$2 RETURNING id,title,status::text,rejection_reason,created_at`, session.Organization.ID, u.ResumeID).Scan(&result.ID, &result.FileName, &result.Status, &result.RejectionReason, &result.CreatedAt); err != nil {
		return model.Resume{}, err
	}
	payload, _ := json.Marshal(map[string]any{"eventId": newID(), "organizationId": session.Organization.ID, "resumeId": u.ResumeID, "fileObjectId": u.FileObjectID, "occurredAt": time.Now().UTC()})
	if _, err = tx.Exec(ctx, `INSERT INTO outbox_events(id,organization_id,event_type,aggregate_type,aggregate_id,payload) VALUES($1,$2,'resume.upload.completed.v1','resume',$3,$4)`, newID(), session.Organization.ID, u.ResumeID, payload); err != nil {
		return model.Resume{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(id,organization_id,actor_user_id,action,resource_type,resource_id,request_id) VALUES($1,$2,$3,'resume.upload_completed','resume',$4,$5)`, newID(), session.Organization.ID, session.User.ID, u.ResumeID, requestID); err != nil {
		return model.Resume{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return model.Resume{}, err
	}
	return result, nil
}

func (s *Store) ListResumes(ctx context.Context, session model.Session) ([]model.Resume, error) {
	tx, err := s.tenantTx(ctx, session)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `SELECT r.id,fo.original_name,r.status::text,r.rejection_reason,r.created_at FROM resumes r JOIN resume_versions rv ON rv.resume_id=r.id AND rv.organization_id=r.organization_id AND rv.version_number=1 JOIN file_objects fo ON fo.id=rv.file_object_id AND fo.organization_id=rv.organization_id WHERE r.organization_id=$1 AND r.user_id=$2 AND r.deleted_at IS NULL ORDER BY r.created_at DESC`, session.Organization.ID, session.User.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.Resume{}
	for rows.Next() {
		var item model.Resume
		if err := rows.Scan(&item.ID, &item.FileName, &item.Status, &item.RejectionReason, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return items, nil
}

func ProfileCompletion(profile model.CareerProfile) int {
	score := 0
	if profile.Headline != "" {
		score += 20
	}
	if profile.Summary != "" {
		score += 15
	}
	if profile.CountryCode != "" && profile.TimeZone != "" {
		score += 15
	}
	if len(profile.Skills) >= 3 {
		score += 20
	}
	if len(profile.Experiences) > 0 {
		score += 20
	}
	if len(profile.Education) > 0 {
		score += 10
	}
	if score > 100 {
		return 100
	}
	return score
}

func NormalizeProfile(profile *model.CareerProfile) {
	profile.Headline = strings.TrimSpace(profile.Headline)
	profile.Summary = strings.TrimSpace(profile.Summary)
	profile.CountryCode = strings.ToUpper(strings.TrimSpace(profile.CountryCode))
	profile.City = strings.TrimSpace(profile.City)
	profile.TimeZone = strings.TrimSpace(profile.TimeZone)
	for i := range profile.Skills {
		profile.Skills[i] = strings.TrimSpace(profile.Skills[i])
	}
}
