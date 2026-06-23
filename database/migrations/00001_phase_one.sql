-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE organization_type AS ENUM ('personal', 'team');
CREATE TYPE membership_role AS ENUM ('owner', 'admin', 'member');
CREATE TYPE onboarding_status AS ENUM ('not_started', 'in_progress', 'completed');
CREATE TYPE profile_method AS ENUM ('resume', 'manual');
CREATE TYPE resume_status AS ENUM ('pending', 'uploaded', 'scanning', 'ready', 'rejected', 'failed');
CREATE TYPE file_scan_status AS ENUM ('pending', 'clean', 'infected', 'invalid', 'failed');
CREATE TYPE resume_source AS ENUM ('original', 'manual', 'generated');

CREATE TABLE users (
  id uuid PRIMARY KEY,
  email citext NOT NULL,
  display_name text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE TABLE auth_identities (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  issuer text NOT NULL,
  subject text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (issuer, subject)
);

CREATE TABLE organizations (
  id uuid PRIMARY KEY,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
  slug text NOT NULL CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
  type organization_type NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  UNIQUE (slug)
);

CREATE TABLE organization_memberships (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role membership_role NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (organization_id, user_id),
  UNIQUE (organization_id, id)
);

CREATE TABLE organization_subscriptions (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
  plan_code text NOT NULL DEFAULT 'free',
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'trialing', 'past_due', 'canceled')),
  provider text,
  provider_customer_id text,
  provider_subscription_id text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (organization_id, id)
);

CREATE TABLE onboarding_progress (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  current_step smallint NOT NULL DEFAULT 1 CHECK (current_step BETWEEN 1 AND 4),
  status onboarding_status NOT NULL DEFAULT 'not_started',
  profile_method profile_method,
  completed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (organization_id, user_id),
  UNIQUE (organization_id, id)
);

CREATE TABLE career_profiles (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  headline text NOT NULL DEFAULT '' CHECK (char_length(headline) <= 160),
  summary text NOT NULL DEFAULT '' CHECK (char_length(summary) <= 4000),
  country_code varchar(2) NOT NULL DEFAULT '' CHECK (country_code = '' OR country_code ~ '^[A-Z]{2}$'),
  time_zone text NOT NULL DEFAULT 'UTC',
  city text NOT NULL DEFAULT '' CHECK (char_length(city) <= 120),
  years_experience numeric(4,1) CHECK (years_experience BETWEEN 0 AND 80),
  version integer NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  UNIQUE (organization_id, user_id),
  UNIQUE (organization_id, id)
);

CREATE TABLE profile_skills (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL,
  career_profile_id uuid NOT NULL,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
  normalized_name text NOT NULL CHECK (char_length(normalized_name) BETWEEN 1 AND 80),
  position smallint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, career_profile_id) REFERENCES career_profiles(organization_id, id) ON DELETE CASCADE,
  UNIQUE (career_profile_id, normalized_name),
  UNIQUE (organization_id, id)
);

CREATE TABLE work_experiences (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL,
  career_profile_id uuid NOT NULL,
  company text NOT NULL CHECK (char_length(company) BETWEEN 1 AND 160),
  title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 160),
  location text NOT NULL DEFAULT '' CHECK (char_length(location) <= 160),
  start_date date NOT NULL,
  end_date date,
  is_current boolean NOT NULL DEFAULT false,
  description text NOT NULL DEFAULT '' CHECK (char_length(description) <= 5000),
  position smallint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, career_profile_id) REFERENCES career_profiles(organization_id, id) ON DELETE CASCADE,
  CHECK ((is_current AND end_date IS NULL) OR (NOT is_current)),
  CHECK (end_date IS NULL OR end_date >= start_date),
  UNIQUE (organization_id, id)
);

CREATE TABLE educations (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL,
  career_profile_id uuid NOT NULL,
  institution text NOT NULL CHECK (char_length(institution) BETWEEN 1 AND 200),
  degree text NOT NULL DEFAULT '' CHECK (char_length(degree) <= 160),
  field_of_study text NOT NULL DEFAULT '' CHECK (char_length(field_of_study) <= 160),
  start_date date,
  end_date date,
  position smallint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, career_profile_id) REFERENCES career_profiles(organization_id, id) ON DELETE CASCADE,
  CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date),
  UNIQUE (organization_id, id)
);

CREATE TABLE projects (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL,
  career_profile_id uuid NOT NULL,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 160),
  description text NOT NULL DEFAULT '' CHECK (char_length(description) <= 3000),
  url text NOT NULL DEFAULT '' CHECK (char_length(url) <= 500),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, career_profile_id) REFERENCES career_profiles(organization_id, id) ON DELETE CASCADE,
  UNIQUE (organization_id, id)
);

CREATE TABLE certifications (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL,
  career_profile_id uuid NOT NULL,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 200),
  issuer text NOT NULL DEFAULT '' CHECK (char_length(issuer) <= 200),
  issued_on date,
  expires_on date,
  credential_url text NOT NULL DEFAULT '' CHECK (char_length(credential_url) <= 500),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, career_profile_id) REFERENCES career_profiles(organization_id, id) ON DELETE CASCADE,
  UNIQUE (organization_id, id)
);

CREATE TABLE file_objects (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  bucket text NOT NULL,
  object_key text NOT NULL,
  original_name text NOT NULL CHECK (char_length(original_name) BETWEEN 1 AND 255),
  content_type text NOT NULL,
  size_bytes bigint NOT NULL CHECK (size_bytes BETWEEN 1 AND 10485760),
  checksum_sha256 char(64) NOT NULL CHECK (checksum_sha256 ~ '^[0-9a-f]{64}$'),
  scan_status file_scan_status NOT NULL DEFAULT 'pending',
  scan_detail text,
  created_at timestamptz NOT NULL DEFAULT now(),
  scanned_at timestamptz,
  deleted_at timestamptz,
  UNIQUE (bucket, object_key),
  UNIQUE (organization_id, id)
);

CREATE TABLE resumes (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title text NOT NULL DEFAULT 'Original resume' CHECK (char_length(title) BETWEEN 1 AND 160),
  status resume_status NOT NULL DEFAULT 'pending',
  rejection_reason text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  UNIQUE (organization_id, id)
);

CREATE TABLE resume_versions (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL,
  resume_id uuid NOT NULL,
  parent_version_id uuid,
  file_object_id uuid,
  source resume_source NOT NULL,
  version_number integer NOT NULL CHECK (version_number > 0),
  structured_data jsonb,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, resume_id) REFERENCES resumes(organization_id, id) ON DELETE CASCADE,
  FOREIGN KEY (organization_id, file_object_id) REFERENCES file_objects(organization_id, id),
  FOREIGN KEY (organization_id, parent_version_id) REFERENCES resume_versions(organization_id, id),
  UNIQUE (resume_id, version_number),
  UNIQUE (organization_id, id)
);

CREATE TABLE resume_uploads (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL,
  resume_id uuid NOT NULL,
  file_object_id uuid NOT NULL,
  expires_at timestamptz NOT NULL,
  completed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (organization_id, resume_id) REFERENCES resumes(organization_id, id) ON DELETE CASCADE,
  FOREIGN KEY (organization_id, file_object_id) REFERENCES file_objects(organization_id, id) ON DELETE CASCADE,
  UNIQUE (organization_id, id)
);

CREATE TABLE outbox_events (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  event_type text NOT NULL,
  aggregate_type text NOT NULL,
  aggregate_id uuid NOT NULL,
  payload jsonb NOT NULL,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  published_at timestamptz,
  attempts integer NOT NULL DEFAULT 0,
  last_error text,
  UNIQUE (organization_id, id)
);

CREATE TABLE audit_logs (
  id uuid PRIMARY KEY,
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  actor_user_id uuid REFERENCES users(id),
  action text NOT NULL,
  resource_type text NOT NULL,
  resource_id uuid,
  request_id text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (organization_id, id)
);

CREATE TABLE idempotency_keys (
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  key text NOT NULL,
  request_hash char(64) NOT NULL,
  response_status integer,
  response_body jsonb,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (organization_id, key)
);

CREATE INDEX idx_memberships_user ON organization_memberships(user_id);
CREATE INDEX idx_experiences_profile ON work_experiences(career_profile_id, position);
CREATE INDEX idx_educations_profile ON educations(career_profile_id, position);
CREATE INDEX idx_skills_profile ON profile_skills(career_profile_id, position);
CREATE INDEX idx_resumes_org_user ON resumes(organization_id, user_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_outbox_unpublished ON outbox_events(occurred_at) WHERE published_at IS NULL;
CREATE INDEX idx_audit_org_time ON audit_logs(organization_id, occurred_at DESC);

ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE onboarding_progress ENABLE ROW LEVEL SECURITY;
ALTER TABLE career_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE profile_skills ENABLE ROW LEVEL SECURITY;
ALTER TABLE work_experiences ENABLE ROW LEVEL SECURITY;
ALTER TABLE educations ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE certifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE file_objects ENABLE ROW LEVEL SECURITY;
ALTER TABLE resumes ENABLE ROW LEVEL SECURITY;
ALTER TABLE resume_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE resume_uploads ENABLE ROW LEVEL SECURITY;
ALTER TABLE outbox_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE idempotency_keys ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_organizations ON organizations USING (id = nullif(current_setting('app.organization_id', true), '')::uuid) WITH CHECK (id = nullif(current_setting('app.organization_id', true), '')::uuid);
CREATE POLICY member_self_access ON organization_memberships USING (organization_id = nullif(current_setting('app.organization_id', true), '')::uuid OR user_id = nullif(current_setting('app.user_id', true), '')::uuid) WITH CHECK (organization_id = nullif(current_setting('app.organization_id', true), '')::uuid);

-- +goose StatementBegin
DO $$
DECLARE table_name text;
BEGIN
  FOREACH table_name IN ARRAY ARRAY['organization_subscriptions','onboarding_progress','career_profiles','profile_skills','work_experiences','educations','projects','certifications','file_objects','resumes','resume_versions','resume_uploads','outbox_events','audit_logs','idempotency_keys']
  LOOP
    EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (organization_id = nullif(current_setting(''app.organization_id'', true), '''')::uuid) WITH CHECK (organization_id = nullif(current_setting(''app.organization_id'', true), '''')::uuid)', table_name);
  END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys, audit_logs, outbox_events, resume_uploads, resume_versions, resumes, file_objects, certifications, projects, educations, work_experiences, profile_skills, career_profiles, onboarding_progress, organization_subscriptions, organization_memberships, organizations, auth_identities, users CASCADE;
DROP TYPE IF EXISTS resume_source, file_scan_status, resume_status, profile_method, onboarding_status, membership_role, organization_type;
