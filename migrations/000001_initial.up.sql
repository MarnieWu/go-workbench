-- 内部用户身份
CREATE TABLE owners(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  oidc_issuer text NOT NULL,
  oidc_subject text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_owner_oidc_issuer_oidc_subject_key UNIQUE (oidc_issuer, oidc_subject),
  CONSTRAINT chk_owner_oidc_issuer CHECK (char_length(oidc_issuer) > 0),
  CONSTRAINT chk_owner_oidc_subject CHECK (char_length(oidc_subject) > 0)
);

-- 用户的项目
CREATE TABLE projects(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  owner_id uuid NOT NULL REFERENCES owners(id),
  name text NOT NULL,
  description text,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_project_owner_id_id UNIQUE (owner_id, id),
  CONSTRAINT chk_project_name CHECK (char_length(name) > 0),
  CONSTRAINT chk_project_status CHECK (status IN ('active', 'archived'))
);

-- 用户的任务
CREATE TABLE tasks(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  owner_id uuid NOT NULL REFERENCES owners(id),
  project_id uuid,
  title text NOT NULL,
  description text,
  status text NOT NULL DEFAULT 'backlog',
  priority text NOT NULL DEFAULT 'none',
  labels text[] NOT NULL DEFAULT '{}',
  due_at timestamptz,
  archived_at timestamptz,
  version bigint NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_task_owner_id_id UNIQUE (owner_id, id),
  CONSTRAINT fk_task_project FOREIGN KEY (owner_id, project_id) REFERENCES projects(owner_id, id),
  CONSTRAINT chk_task_title CHECK (char_length(title) > 0),
  CONSTRAINT chk_task_status CHECK (status IN ('backlog', 'in_progress', 'blocked', 'done')),
  CONSTRAINT chk_task_priority CHECK (priority IN ('none', 'low', 'medium', 'high')),
  CONSTRAINT chk_task_version CHECK (version > 0)
);

-- 原始输入及处理进度
CREATE TABLE captures(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  owner_id uuid NOT NULL REFERENCES owners(id),
  idempotency_key text NOT NULL,
  input_hash text NOT NULL,
  input_text text NOT NULL,
  source_type text NOT NULL,
  status text NOT NULL DEFAULT 'queued',
  attempt_count integer NOT NULL DEFAULT 0,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_capture_owner_id_idempotency_key UNIQUE (owner_id, idempotency_key),
  CONSTRAINT uk_capture_owner_id_id UNIQUE (owner_id, id),
  CONSTRAINT chk_capture_idempotency_key CHECK (char_length(idempotency_key) > 0),
  CONSTRAINT chk_capture_input_hash CHECK (input_hash ~ '^[0-9a-f]{64}$'),
  CONSTRAINT chk_capture_input_text CHECK (char_length(input_text) > 0 AND char_length(input_text) <= 250),
  CONSTRAINT chk_status CHECK (status IN ('queued', 'processing', 'processed', 'failed')),
  CONSTRAINT chk_capture_attempt_count CHECK (attempt_count >= 0)
);

-- 用户的检查项
CREATE TABLE check_items(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  task_id uuid NOT NULL REFERENCES tasks(id),
  content text NOT NULL,
  completed boolean NOT NULL DEFAULT FALSE,
  position integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_check_item_task_id_position UNIQUE (task_id, position),
  CONSTRAINT chk_check_item_content CHECK (char_length(content) > 0),
  CONSTRAINT chk_check_item_position CHECK (position >= 0)
);

-- 等待审核的建议
CREATE TABLE candidates(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  owner_id uuid NOT NULL,
  capture_id uuid NOT NULL,
  proposed_title text NOT NULL,
  proposed_description text,
  proposed_project_id uuid,
  labels text[] NOT NULL DEFAULT '{}',
  status text NOT NULL DEFAULT 'pending_review',
  accepted_task_id uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_candidate_task_id UNIQUE (accepted_task_id),
  CONSTRAINT fk_candidate_owner_id_proposed_project_id FOREIGN KEY (owner_id, proposed_project_id) REFERENCES projects(owner_id, id),
  CONSTRAINT fk_candidate_owner_id_task_id FOREIGN KEY (owner_id, accepted_task_id) REFERENCES tasks(owner_id, id),
  CONSTRAINT fk_candidate_capture FOREIGN KEY (owner_id, capture_id) REFERENCES captures(owner_id, id),
  CONSTRAINT chk_proposed_title CHECK (char_length(proposed_title) > 0),
  CONSTRAINT chk_status CHECK (status IN ('pending_review', 'accepted', 'rejected'))
);

-- 输入来源的最小证据
CREATE TABLE source_evidence(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  owner_id uuid NOT NULL,
  capture_id uuid NOT NULL,
  source_type text NOT NULL,
  external_ref text,
  source_url text,
  excerpt text,
  consent_scope text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_source_evidence_owner_id_id UNIQUE (owner_id, id),
  CONSTRAINT fk_source_evidence_owner_id_capture_id FOREIGN KEY (owner_id, capture_id) REFERENCES captures(owner_id, id)
);

-- Task 与证据的关联
CREATE TABLE task_evidence_links(
  owner_id uuid NOT NULL,
  task_id uuid NOT NULL,
  source_evidence_id uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_task_evidence_link_task_id_source_evidence_id UNIQUE (task_id, source_evidence_id),
  CONSTRAINT fk_task_evidence_link_owner_id_task_id FOREIGN KEY (owner_id, task_id) REFERENCES tasks(owner_id, id),
  CONSTRAINT fk_task_evidence_link_owner_id_source_evidence_id FOREIGN KEY (owner_id, source_evidence_id) REFERENCES source_evidence(owner_id, id)
);

-- 等待确认的完成或归档请求
CREATE TABLE pending_actions(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  owner_id uuid NOT NULL REFERENCES owners(id),
  target_type text NOT NULL,
  target_id uuid NOT NULL,
  target_version bigint NOT NULL,
  action text NOT NULL,
  parameters jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'pending',
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_target_version CHECK (target_version > 0),
  CONSTRAINT chk_pending_parameters_object CHECK (jsonb_typeof(parameters) = 'object'),
  CONSTRAINT chk_status CHECK (status IN ('pending', 'executed', 'cancelled', 'expired', 'failed')),
  CONSTRAINT chk_expires_at CHECK (expires_at > created_at)
);

-- 操作记录
CREATE TABLE audit_events(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  owner_id uuid NOT NULL REFERENCES owners(id),
  actor_type text NOT NULL,
  actor_id text NOT NULL,
  action text NOT NULL,
  entity_type text NOT NULL,
  entity_id text NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  request_id text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_audit_metadata_object CHECK (jsonb_typeof(metadata) = 'object')
);

