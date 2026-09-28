-- 内部用户身份
CREATE TABLE
  owners (
    id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
    oidc_issuer text NOT NULL,
    oidc_subject text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now (),
    CONSTRAINT uk_oidc_issuer_oidc_subject_key UNIQUE (oidc_issuer, oidc_subject),
    CONSTRAINT chk_owner_oidc_issuer CHECK char_length(oidc_issuer) > 0,
    CONSTRAINT chk_owner_oidc_subject CHECK char_length(oidc_subject) > 0,
  )
  -- 用户的项目
CREATE TABLE
  projects (
    id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
    owner_id text NOT NULL REFERENCES owners (id),
    name text NOT NULL,
    description text,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    CONSTRAINT chk_project_name CHECK char_length(name) > 0,
    CONSTRAINT chk_project_status CHECK status IN ('active', 'archived'),
  )
  -- 用户的任务
CREATE TABLE
  tasks (
    id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
    owner_id text NOT NULL REFERENCES owners (id),
    project_id text,
    title text NOT NULL,
    description text NOT NULL,
    status text NOT NULL DEFAULT 'backlog',
    priority text NOT NULL DEFAULT 'none',
    labels text[] DEFAULT '{}',
    due_at timestamptz,
    archived_at timestamptz,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    CONSTRAINT chk_task_title CHECK char_length(title) > 0,
    CONSTRAINT chk_task_status CHECK status IN ('backlog', 'in_progress', 'blocked', 'done'),
    CONSTRAINT chk_task_priority CHECK priority IN ('none', 'low', 'medium', 'high'),
    CONSTRAINT chk_task_labels CHECK cardinality(labels, 1) > 0,
    CONSTRAINT chk_task_version CHECK version > 0,
  )
  -- 用户的检查项
CREATE TABLE
  check_items (
    id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
    task_id text NOT NULL REFERENCES tasks (id),
    content text NOT NULL,
    completed boolean DEFAULT false,
    position integer NOT NULL DEFAULT 0,
    created_at timestamptz DEFAULT now (),
    updated_at timestamptz DEFAULT now (),
    CONSTRAINT chk_check_item_content CHECK char_length(content) > 0,
    CONSTRAINT chk_check_item_position CHECK position >= 0,
  )
  -- 原始输入及处理进度
CREATE TABLE
  captures (
    id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
    owner_id text NOT NULL REFERENCES owners (id),
    idempotency_key text NOT NULL,
    input_hash text NOT NULL,
    input_text text NOT NULL,
    source_type text NOT NULL,
    status text NOT NULL DEFAULT 'queued',
    attempt_count integer NOT NULL DEFAULT 0,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    CONSTRAINT uk_owner_id_idempotency_key UNIQUE (owner_id, idempotency_key),
    CONSTRAINT chk_capture_idempotency_key CHECK char_length(idempotency_key) > 0,
    CONSTRAINT chk_capture_input_hash CHECK char_length(input_hash) = 64,
    CONSTRAINT chk_capture_input_text CHECK (
      char_length(input_text) > 0
      AND char_length(input_text) <= 250
    ),
    CONSTRAINT chk_status CHECK status IN ('queued', 'processing', 'processed', 'failed'),
    CONSTRAINT chk_capture_attempt_count CHECK attempt_count >= 0,
  )
  -- 等待审核的建议
CREATE TABLE
  candidates (
    id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
    owner_id text NOT NULL REFERENCES owners (id),
    capture_id text NOT NULL REFERENCES captures (id),
    proposed_title text NOT NULL,
    proposed_description text,
    proposed_project_id text,
    labels text[] DEFAULT '{)',
    status text NOT NULL DEFAULT 'pending_review',
    accepted_task_id text,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    CONSTRAINT chk_proposed_title CHECK char_length(proposed_title) > 0,
    CONSTRAINT chk_status CHECK status IN ('pending_review', 'accepted', 'rejected'),
  )
  -- 输入来源的最小证据
CREATE TABLE
  source_evidence (
    id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
    owner_id text NOT NULL REFERENCES owners (id),
    capture_id text NOT NULL REFERENCES captures (id),
    source_type text NOT NULL REFERENCES captures (source_type),
    external_ref text,
    source_url text,
    excerpt text,
    consent_scope text,
    created_at timestamptz NOT NULL DEFAULT now (),
  )
  -- Task 与证据的关联
CREATE TABLE
  task_evidence_links (
    owner_id text NOT NULL REFERENCES owners (id),
    task_id text NOT NULL REFERENCES tasks (id),
    source_evidence_id text NOT NULL REFERENCES source_evidence (id),
    created_at timestamptz NOT NULL DEFAULT now (),
    CONSTRAINT uk_task_id_source_evidence_id UNIQUE (task_id, source_evidence_id),
  )
  -- 等待确认的完成或归档请求
CREATE TABLE
  pending_actions (
    id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
    owner_id text NOT NULL REFERENCES owners (id),
    target_type text NOT NULL,
    target_id text NOT NULL,
    target_version bigint NOT NULL,
    action text,
    parameters jsonb NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now (),
    updated_at timestamptz NOT NULL DEFAULT now (),
    CONSTRAINT chk_target_version CHECK target_version > 0,
    CONSTRAINT chk_status CHECK status IN (
      'pending',
      'executed',
      'cancelled',
      'expired',
      'failed'
    ),
    CONSTRAINT chk_expires_at CHECK (expires_at > created_at),
  )
  -- 操作记录
CREATE TABLE
  audit_events (
    id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
    owner_id text NOT NULL REFERENCES owners (id),
    actor_type text NOT NULL,
    actor_id text NOT NULL,
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id text NOT NULL,
    metadata jsonb NOT NULL,
    request_id text,
    created_at timestamptz NOT NULL DEFAULT now (),
  )