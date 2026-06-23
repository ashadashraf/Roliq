package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/roliq/roliq/internal/auth"
	"github.com/roliq/roliq/internal/model"
	"github.com/roliq/roliq/internal/observability"
	"github.com/roliq/roliq/internal/storage"
	"github.com/roliq/roliq/internal/store"
)

const (
	claimsKey  = "authClaims"
	sessionKey = "tenantSession"
)

type Server struct {
	echo      *echo.Echo
	store     *store.Store
	objects   storage.Store
	verifier  auth.Verifier
	redis     *redis.Client
	bucket    string
	webOrigin string
	logger    *slog.Logger
}

func New(st *store.Store, objects storage.Store, verifier auth.Verifier, redisClient *redis.Client, bucket, webOrigin string, logger *slog.Logger) *Server {
	s := &Server{store: st, objects: objects, verifier: verifier, redis: redisClient, bucket: bucket, webOrigin: webOrigin, logger: logger}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = s.errorHandler
	e.Use(middleware.RequestID(), middleware.Recover(), middleware.BodyLimit("12M"))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{XSSProtection: "0", ContentTypeNosniff: "nosniff", XFrameOptions: "DENY", ReferrerPolicy: "strict-origin-when-cross-origin"}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{webOrigin}, AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions}, AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Organization-ID", "Idempotency-Key", "X-Request-ID"}, ExposeHeaders: []string{"X-Request-ID"}, MaxAge: 3600}))
	e.Use(s.requestLogger, s.rateLimit)
	e.Use(observability.MetricsMiddleware, observability.TracingMiddleware)
	e.GET("/health/live", func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]string{"status": "ok"}) })
	e.GET("/health/ready", s.ready)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	v1 := e.Group("/v1", s.authenticate)
	v1.POST("/session/bootstrap", s.bootstrap)
	tenant := v1.Group("", s.loadSession)
	tenant.GET("/me", s.me)
	tenant.GET("/dashboard", s.dashboard)
	tenant.GET("/onboarding", s.getOnboarding)
	tenant.PATCH("/onboarding", s.updateOnboarding)
	tenant.GET("/career-profile", s.getProfile)
	tenant.PUT("/career-profile", s.saveProfile)
	tenant.GET("/resumes", s.listResumes)
	tenant.POST("/resume-uploads", s.createUpload)
	tenant.POST("/resume-uploads/:id/complete", s.completeUpload)
	s.echo = e
	return s
}

func (s *Server) Echo() *echo.Echo { return s.echo }

func (s *Server) requestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		status := c.Response().Status
		if err != nil {
			var he *echo.HTTPError
			if errors.As(err, &he) {
				status = he.Code
			} else {
				status = 500
			}
		}
		s.logger.InfoContext(c.Request().Context(), "http_request", "method", c.Request().Method, "path", c.Path(), "status", status, "duration_ms", time.Since(start).Milliseconds(), "request_id", c.Response().Header().Get(echo.HeaderXRequestID))
		return err
	}
}

func (s *Server) rateLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if s.redis == nil {
			return next(c)
		}
		now := time.Now().UTC()
		key := fmt.Sprintf("rate:%s:%s", c.RealIP(), now.Format("200601021504"))
		count, err := s.redis.Incr(c.Request().Context(), key).Result()
		if err != nil {
			s.logger.Warn("rate_limit_unavailable", "error", err)
			return next(c)
		}
		if count == 1 {
			s.redis.Expire(c.Request().Context(), key, 2*time.Minute)
		}
		if count > 180 {
			return problem(c, http.StatusTooManyRequests, "rate_limit_exceeded", "Too many requests", "Please wait before trying again.", nil)
		}
		return next(c)
	}
}

func (s *Server) authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get(echo.HeaderAuthorization)
		if !strings.HasPrefix(header, "Bearer ") {
			return problem(c, http.StatusUnauthorized, "authentication_required", "Authentication required", "A valid bearer token is required.", nil)
		}
		claims, err := s.verifier.Verify(c.Request().Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			s.logger.Info("authentication_rejected", "error", err)
			return problem(c, http.StatusUnauthorized, "invalid_token", "Authentication failed", "The access token is invalid or expired.", nil)
		}
		c.Set(claimsKey, claims)
		return next(c)
	}
}

func (s *Server) loadSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := s.store.ResolveSession(c.Request().Context(), claimsFrom(c), c.Request().Header.Get("X-Organization-ID"))
		if errors.Is(err, store.ErrNotFound) {
			return problem(c, http.StatusForbidden, "workspace_access_denied", "Workspace unavailable", "The requested workspace is unavailable.", nil)
		}
		if err != nil {
			return err
		}
		c.Set(sessionKey, session)
		return next(c)
	}
}
func claimsFrom(c echo.Context) auth.Claims    { return c.Get(claimsKey).(auth.Claims) }
func sessionFrom(c echo.Context) model.Session { return c.Get(sessionKey).(model.Session) }

func (s *Server) ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		return problem(c, http.StatusServiceUnavailable, "dependency_unavailable", "Service unavailable", "A required dependency is unavailable.", nil)
	}
	if err := s.redis.Ping(ctx).Err(); err != nil {
		return problem(c, http.StatusServiceUnavailable, "dependency_unavailable", "Service unavailable", "A required dependency is unavailable.", nil)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) bootstrap(c echo.Context) error {
	session, err := s.store.Bootstrap(c.Request().Context(), claimsFrom(c), requestID(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, session)
}
func (s *Server) me(c echo.Context) error { return c.JSON(http.StatusOK, sessionFrom(c)) }
func (s *Server) getOnboarding(c echo.Context) error {
	return c.JSON(http.StatusOK, sessionFrom(c).Onboarding)
}

type onboardingPatch struct {
	CurrentStep   *int    `json:"currentStep"`
	Status        *string `json:"status"`
	ProfileMethod *string `json:"profileMethod"`
}

func (s *Server) updateOnboarding(c echo.Context) error {
	var input onboardingPatch
	if err := decodeJSON(c, &input); err != nil {
		return invalidJSON(c, err)
	}
	fields := map[string]string{}
	if input.CurrentStep != nil && (*input.CurrentStep < 1 || *input.CurrentStep > 4) {
		fields["currentStep"] = "Must be between 1 and 4."
	}
	if input.Status != nil && !oneOf(*input.Status, "not_started", "in_progress", "completed") {
		fields["status"] = "Invalid onboarding status."
	}
	if input.ProfileMethod != nil && !oneOf(*input.ProfileMethod, "resume", "manual") {
		fields["profileMethod"] = "Choose resume or manual."
	}
	if len(fields) > 0 {
		return problem(c, 422, "validation_failed", "Validation failed", "Review the highlighted fields.", fields)
	}
	result, err := s.store.UpdateOnboarding(c.Request().Context(), sessionFrom(c), input.CurrentStep, input.Status, input.ProfileMethod, requestID(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (s *Server) getProfile(c echo.Context) error {
	profile, err := s.store.GetProfile(c.Request().Context(), sessionFrom(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, profile)
}
func (s *Server) saveProfile(c echo.Context) error {
	var input model.CareerProfile
	if err := decodeJSON(c, &input); err != nil {
		return invalidJSON(c, err)
	}
	store.NormalizeProfile(&input)
	if fields := validateProfile(input); len(fields) > 0 {
		return problem(c, 422, "validation_failed", "Validation failed", "Review the highlighted fields.", fields)
	}
	result, err := s.store.SaveProfile(c.Request().Context(), sessionFrom(c), input, requestID(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func validateProfile(p model.CareerProfile) map[string]string {
	fields := map[string]string{}
	if len(p.Headline) > 160 {
		fields["headline"] = "Use 160 characters or fewer."
	}
	if len(p.Summary) > 4000 {
		fields["summary"] = "Use 4,000 characters or fewer."
	}
	if p.CountryCode != "" {
		valid := len(p.CountryCode) == 2 && p.CountryCode[0] >= 'A' && p.CountryCode[0] <= 'Z' && p.CountryCode[1] >= 'A' && p.CountryCode[1] <= 'Z'
		if !valid {
			fields["countryCode"] = "Use a two-letter ISO country code."
		}
	}
	if _, err := time.LoadLocation(p.TimeZone); err != nil {
		fields["timeZone"] = "Use a valid IANA time zone."
	}
	if p.YearsExperience != nil && (*p.YearsExperience < 0 || *p.YearsExperience > 80) {
		fields["yearsExperience"] = "Must be between 0 and 80."
	}
	if len(p.Skills) > 100 {
		fields["skills"] = "Add no more than 100 skills."
	}
	for i, skill := range p.Skills {
		if length := len([]rune(strings.TrimSpace(skill))); length == 0 || length > 80 {
			fields[fmt.Sprintf("skills.%d", i)] = "Each skill must contain 1 to 80 characters."
		}
	}
	if len(p.Experiences) > 100 {
		fields["experiences"] = "Add no more than 100 roles."
	}
	for i, item := range p.Experiences {
		if strings.TrimSpace(item.Company) == "" {
			fields[fmt.Sprintf("experiences.%d.company", i)] = "Company is required."
		}
		if strings.TrimSpace(item.Title) == "" {
			fields[fmt.Sprintf("experiences.%d.title", i)] = "Title is required."
		}
		if len([]rune(item.Company)) > 160 {
			fields[fmt.Sprintf("experiences.%d.company", i)] = "Use 160 characters or fewer."
		}
		if len([]rune(item.Title)) > 160 {
			fields[fmt.Sprintf("experiences.%d.title", i)] = "Use 160 characters or fewer."
		}
		if len([]rune(item.Location)) > 160 {
			fields[fmt.Sprintf("experiences.%d.location", i)] = "Use 160 characters or fewer."
		}
		if len([]rune(item.Description)) > 5000 {
			fields[fmt.Sprintf("experiences.%d.description", i)] = "Use 5,000 characters or fewer."
		}
		start, err := time.Parse("2006-01-02", item.StartDate)
		if err != nil {
			fields[fmt.Sprintf("experiences.%d.startDate", i)] = "Use YYYY-MM-DD."
		}
		if item.EndDate != nil {
			end, e := time.Parse("2006-01-02", *item.EndDate)
			if e != nil || (!start.IsZero() && end.Before(start)) {
				fields[fmt.Sprintf("experiences.%d.endDate", i)] = "End date must follow the start date."
			}
		}
		if item.IsCurrent && item.EndDate != nil {
			fields[fmt.Sprintf("experiences.%d.endDate", i)] = "Current roles cannot have an end date."
		}
	}
	if len(p.Education) > 100 {
		fields["education"] = "Add no more than 100 education records."
	}
	for i, item := range p.Education {
		if strings.TrimSpace(item.Institution) == "" {
			fields[fmt.Sprintf("education.%d.institution", i)] = "Institution is required."
		}
		if len([]rune(item.Institution)) > 200 {
			fields[fmt.Sprintf("education.%d.institution", i)] = "Use 200 characters or fewer."
		}
		if len([]rune(item.Degree)) > 160 {
			fields[fmt.Sprintf("education.%d.degree", i)] = "Use 160 characters or fewer."
		}
		if len([]rune(item.FieldOfStudy)) > 160 {
			fields[fmt.Sprintf("education.%d.fieldOfStudy", i)] = "Use 160 characters or fewer."
		}
		if item.StartDate != nil {
			if _, err := time.Parse("2006-01-02", *item.StartDate); err != nil {
				fields[fmt.Sprintf("education.%d.startDate", i)] = "Use YYYY-MM-DD."
			}
		}
		if item.EndDate != nil {
			end, endErr := time.Parse("2006-01-02", *item.EndDate)
			if endErr != nil {
				fields[fmt.Sprintf("education.%d.endDate", i)] = "Use YYYY-MM-DD."
			}
			if item.StartDate != nil {
				if start, err := time.Parse("2006-01-02", *item.StartDate); err == nil && endErr == nil && end.Before(start) {
					fields[fmt.Sprintf("education.%d.endDate", i)] = "End date must follow the start date."
				}
			}
		}
	}
	return fields
}

type createUploadRequest struct {
	FileName       string `json:"fileName"`
	ContentType    string `json:"contentType"`
	SizeBytes      int64  `json:"sizeBytes"`
	ChecksumSHA256 string `json:"checksumSha256"`
}
type uploadIntent struct {
	UploadID        string            `json:"uploadId"`
	ResumeID        string            `json:"resumeId"`
	UploadURL       string            `json:"uploadUrl"`
	ObjectKey       string            `json:"objectKey"`
	ExpiresAt       time.Time         `json:"expiresAt"`
	RequiredHeaders map[string]string `json:"requiredHeaders"`
}

func (s *Server) createUpload(c echo.Context) error {
	var input createUploadRequest
	if err := decodeJSON(c, &input); err != nil {
		return invalidJSON(c, err)
	}
	input.FileName = strings.TrimSpace(filepath.Base(input.FileName))
	input.ChecksumSHA256 = strings.ToLower(strings.TrimSpace(input.ChecksumSHA256))
	fields := map[string]string{}
	allowed := map[string]string{"application/pdf": ".pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx"}
	ext, ok := allowed[input.ContentType]
	if !ok {
		fields["contentType"] = "Upload a PDF or DOCX file."
	}
	if input.SizeBytes < 1 || input.SizeBytes > 10*1024*1024 {
		fields["sizeBytes"] = "File size must be between 1 byte and 10 MB."
	}
	if len(input.FileName) < 1 || len(input.FileName) > 255 || strings.ToLower(filepath.Ext(input.FileName)) != ext {
		fields["fileName"] = "The file name and content type do not match."
	}
	raw, err := hex.DecodeString(input.ChecksumSHA256)
	if err != nil || len(raw) != sha256.Size {
		fields["checksumSha256"] = "Provide a lowercase SHA-256 checksum."
	}
	if len(fields) > 0 {
		return problem(c, 422, "validation_failed", "Upload rejected", "Review the file details.", fields)
	}
	idempotencyKey := strings.TrimSpace(c.Request().Header.Get("Idempotency-Key"))
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return problem(c, 400, "idempotency_key_required", "Idempotency key required", "Provide an Idempotency-Key header between 16 and 128 characters.", nil)
	}
	requestDigest := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%d\x00%s", input.FileName, input.ContentType, input.SizeBytes, input.ChecksumSHA256)))
	requestHash := hex.EncodeToString(requestDigest[:])
	session := sessionFrom(c)
	resumeID := uuid.Must(uuid.NewV7()).String()
	objectKey := fmt.Sprintf("quarantine/%s/%s/%s%s", session.Organization.ID, resumeID, uuid.Must(uuid.NewV7()).String(), ext)
	expires := time.Now().UTC().Add(15 * time.Minute)
	upload, err := s.store.CreateUpload(c.Request().Context(), session, input.FileName, input.ContentType, input.ChecksumSHA256, s.bucket, objectKey, input.SizeBytes, expires, requestID(c), idempotencyKey, requestHash)
	if errors.Is(err, store.ErrIdempotencyConflict) {
		return problem(c, 409, "idempotency_conflict", "Request conflict", "The Idempotency-Key was already used for a different upload.", nil)
	}
	if err != nil {
		return err
	}
	presignTTL := time.Until(upload.ExpiresAt)
	if presignTTL <= 0 {
		return problem(c, 409, "upload_expired", "Upload expired", "Create a new upload session.", nil)
	}
	url, err := s.objects.PresignUpload(c.Request().Context(), upload.ObjectKey, upload.ContentType, upload.Checksum, presignTTL)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, uploadIntent{UploadID: upload.ID, ResumeID: upload.ResumeID, UploadURL: url, ObjectKey: upload.ObjectKey, ExpiresAt: upload.ExpiresAt, RequiredHeaders: map[string]string{"Content-Type": upload.ContentType, "x-amz-meta-sha256": upload.Checksum}})
}

func (s *Server) completeUpload(c echo.Context) error {
	if _, err := uuid.Parse(c.Param("id")); err != nil {
		return problem(c, 404, "upload_not_found", "Upload not found", "The upload session was not found.", nil)
	}
	session := sessionFrom(c)
	upload, err := s.store.GetUpload(c.Request().Context(), session, c.Param("id"))
	if errors.Is(err, store.ErrNotFound) {
		return problem(c, 404, "upload_not_found", "Upload not found", "The upload session was not found.", nil)
	}
	if err != nil {
		return err
	}
	info, err := s.objects.Head(c.Request().Context(), upload.ObjectKey)
	if err != nil {
		return problem(c, 409, "object_not_available", "Upload incomplete", "The uploaded object could not be verified.", nil)
	}
	if info.SizeBytes != upload.SizeBytes || info.ContentType != upload.ContentType || !strings.EqualFold(info.SHA256, upload.Checksum) {
		return problem(c, 422, "object_verification_failed", "Upload verification failed", "The uploaded object does not match the declared file metadata.", nil)
	}
	resume, err := s.store.CompleteUpload(c.Request().Context(), session, upload, requestID(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resume)
}

func (s *Server) listResumes(c echo.Context) error {
	items, err := s.store.ListResumes(c.Request().Context(), sessionFrom(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}
func (s *Server) dashboard(c echo.Context) error {
	session := sessionFrom(c)
	profile, err := s.store.GetProfile(c.Request().Context(), session)
	if err != nil {
		return err
	}
	resumes, err := s.store.ListResumes(c.Request().Context(), session)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.Dashboard{ProfileCompletion: store.ProfileCompletion(profile), Onboarding: session.Onboarding, Profile: &profile, Resumes: resumes})
}

func decodeJSON(c echo.Context, target any) error {
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("request body must contain one JSON object")
		}
		return err
	}
	return nil
}
func invalidJSON(c echo.Context, err error) error {
	return problem(c, 400, "invalid_json", "Invalid request", "The JSON request body is invalid: "+err.Error(), nil)
}
func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}
func requestID(c echo.Context) string { return c.Response().Header().Get(echo.HeaderXRequestID) }

func problem(c echo.Context, status int, code, title, detail string, fields map[string]string) error {
	body := map[string]any{"type": "https://docs.roliq.com/problems/" + code, "title": title, "status": status, "code": code, "detail": detail, "requestId": requestID(c)}
	if len(fields) > 0 {
		body["fields"] = fields
	}
	return c.Blob(status, "application/problem+json", mustJSON(body))
}
func mustJSON(value any) []byte { data, _ := json.Marshal(value); return data }
func (s *Server) errorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	var he *echo.HTTPError
	if errors.As(err, &he) {
		detail := http.StatusText(he.Code)
		if message, ok := he.Message.(string); ok {
			detail = message
		}
		_ = problem(c, he.Code, "http_error", http.StatusText(he.Code), detail, nil)
		return
	}
	s.logger.ErrorContext(c.Request().Context(), "unhandled_request_error", "error", err, "request_id", requestID(c))
	_ = problem(c, 500, "internal_error", "Internal server error", "The request could not be completed.", nil)
}

func RedisFromURL(value string) (*redis.Client, error) {
	options, err := redis.ParseURL(value)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(options), nil
}
func ParseInt(value string, fallback int) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return result
}
