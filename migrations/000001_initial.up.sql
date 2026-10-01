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
COMMENT ON TABLE owners IS '工作台内部用户与 OIDC 身份的映射';
COMMENT ON COLUMN owners.id IS '内部用户 UUID';
COMMENT ON COLUMN owners.oidc_issuer IS 'OIDC 签发方标识';
COMMENT ON COLUMN owners.oidc_subject IS '签发方内的用户标识';
COMMENT ON COLUMN owners.created_at IS '内部用户创建时间';

-- 用户的项目
CREATE TABLE projects(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  owner_id uuid NOT NULL REFERENCES owners(id),
  name text NOT NULL,
  description text,
  status text NOT NULL DEFAULT 'active',
  version bigint NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_project_owner_id_id UNIQUE (owner_id, id),
  CONSTRAINT chk_project_name CHECK (char_length(name) > 0),
  CONSTRAINT chk_project_status CHECK (status IN ('active', 'archived')),
  CONSTRAINT chk_project_version CHECK (version > 0)
);
COMMENT ON TABLE projects IS '用户创建的任务分组';
COMMENT ON COLUMN projects.id IS '项目 UUID';
COMMENT ON COLUMN projects.owner_id IS '所属内部用户 UUID';
COMMENT ON COLUMN projects.name IS '项目名称';
COMMENT ON COLUMN projects.description IS '项目说明，可为空';
COMMENT ON COLUMN projects.status IS '项目状态：active 或 archived';
COMMENT ON COLUMN projects.version IS '乐观并发版本，归档确认时校验';
COMMENT ON COLUMN projects.created_at IS '项目创建时间';
COMMENT ON COLUMN projects.updated_at IS '项目最后修改时间';

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
COMMENT ON TABLE tasks IS '用户明确创建或接受的正式任务';
COMMENT ON COLUMN tasks.id IS '任务 UUID';
COMMENT ON COLUMN tasks.owner_id IS '所属内部用户 UUID';
COMMENT ON COLUMN tasks.project_id IS '所属项目 UUID，可为空';
COMMENT ON COLUMN tasks.title IS '任务标题';
COMMENT ON COLUMN tasks.description IS '任务描述，可为空';
COMMENT ON COLUMN tasks.status IS '任务进度：backlog、in_progress、blocked 或 done';
COMMENT ON COLUMN tasks.priority IS '任务优先级：none、low、medium 或 high';
COMMENT ON COLUMN tasks.labels IS '任务标签列表，默认空数组';
COMMENT ON COLUMN tasks.due_at IS '任务截止时间，可为空';
COMMENT ON COLUMN tasks.archived_at IS '归档时间，为空表示未归档';
COMMENT ON COLUMN tasks.version IS '乐观并发版本';
COMMENT ON COLUMN tasks.created_at IS '任务创建时间';
COMMENT ON COLUMN tasks.updated_at IS '任务最后修改时间';

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
  CONSTRAINT chk_capture_source_type CHECK (char_length(source_type) > 0),
  CONSTRAINT chk_status CHECK (status IN ('queued', 'processing', 'processed', 'failed')),
  CONSTRAINT chk_capture_attempt_count CHECK (attempt_count >= 0)
);
COMMENT ON TABLE captures IS '一次提交的最小输入及处理进度';
COMMENT ON COLUMN captures.id IS '捕获记录 UUID';
COMMENT ON COLUMN captures.owner_id IS '所属内部用户 UUID';
COMMENT ON COLUMN captures.idempotency_key IS '同一用户下用于精确重试的幂等键';
COMMENT ON COLUMN captures.input_hash IS '输入内容的 SHA-256 小写十六进制摘要';
COMMENT ON COLUMN captures.input_text IS '供 Worker 处理的最小输入文本';
COMMENT ON COLUMN captures.source_type IS '本次提交的来源类别';
COMMENT ON COLUMN captures.status IS '处理状态：queued、processing、processed 或 failed';
COMMENT ON COLUMN captures.attempt_count IS '处理尝试次数';
COMMENT ON COLUMN captures.last_error IS '脱敏后的最近一次失败摘要，可为空';
COMMENT ON COLUMN captures.created_at IS '提交创建时间';
COMMENT ON COLUMN captures.updated_at IS '处理状态最后修改时间';

-- 用户的检查项
CREATE TABLE check_items(
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  task_id uuid NOT NULL REFERENCES tasks(id),
  content text NOT NULL,
  completed boolean NOT NULL DEFAULT FALSE,
  position integer NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uk_check_item_task_id_position UNIQUE (task_id, position),
  CONSTRAINT chk_check_item_content CHECK (char_length(content) > 0),
  CONSTRAINT chk_check_item_position CHECK (position >= 0)
);
COMMENT ON TABLE check_items IS '完成任务所需的检查步骤';
COMMENT ON COLUMN check_items.id IS '检查项 UUID';
COMMENT ON COLUMN check_items.task_id IS '所属任务 UUID';
COMMENT ON COLUMN check_items.content IS '检查项内容';
COMMENT ON COLUMN check_items.completed IS '检查项是否完成，默认否';
COMMENT ON COLUMN check_items.position IS '同一任务内的显示位置，由写入方指定';
COMMENT ON COLUMN check_items.created_at IS '检查项创建时间';
COMMENT ON COLUMN check_items.updated_at IS '检查项最后修改时间';

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
  CONSTRAINT chk_status CHECK (status IN ('pending_review', 'accepted', 'rejected')),
  CONSTRAINT chk_candidate_acceptance CHECK ((status = 'accepted') = (accepted_task_id IS NOT NULL))
);
COMMENT ON TABLE candidates IS '从输入生成、等待用户审核的任务建议';
COMMENT ON COLUMN candidates.id IS '候选任务 UUID';
COMMENT ON COLUMN candidates.owner_id IS '所属内部用户 UUID';
COMMENT ON COLUMN candidates.capture_id IS '产生本建议的捕获记录 UUID';
COMMENT ON COLUMN candidates.proposed_title IS '建议的任务标题';
COMMENT ON COLUMN candidates.proposed_description IS '建议的任务描述，可为空';
COMMENT ON COLUMN candidates.proposed_project_id IS '建议关联的项目 UUID，可为空';
COMMENT ON COLUMN candidates.labels IS '建议的标签列表，默认空数组';
COMMENT ON COLUMN candidates.status IS '审核状态：pending_review、accepted 或 rejected';
COMMENT ON COLUMN candidates.accepted_task_id IS '接受后创建的正式任务 UUID，可为空';
COMMENT ON COLUMN candidates.created_at IS '建议创建时间';
COMMENT ON COLUMN candidates.updated_at IS '建议最后修改时间';

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
  CONSTRAINT fk_source_evidence_owner_id_capture_id FOREIGN KEY (owner_id, capture_id) REFERENCES captures(owner_id, id),
  CONSTRAINT chk_source_evidence_source_type CHECK (char_length(source_type) > 0)
);
COMMENT ON TABLE source_evidence IS '与捕获记录关联的最小来源证据';
COMMENT ON COLUMN source_evidence.id IS '来源证据 UUID';
COMMENT ON COLUMN source_evidence.owner_id IS '所属内部用户 UUID';
COMMENT ON COLUMN source_evidence.capture_id IS '对应的捕获记录 UUID';
COMMENT ON COLUMN source_evidence.source_type IS '该证据的来源类别';
COMMENT ON COLUMN source_evidence.external_ref IS '外部消息或事件标识，可为空';
COMMENT ON COLUMN source_evidence.source_url IS '来源位置链接，可为空';
COMMENT ON COLUMN source_evidence.excerpt IS '说明任务来历的最小摘录，可为空';
COMMENT ON COLUMN source_evidence.consent_scope IS '来源授权范围，可为空';
COMMENT ON COLUMN source_evidence.created_at IS '来源证据创建时间';

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
COMMENT ON TABLE task_evidence_links IS '正式任务与来源证据的关联';
COMMENT ON COLUMN task_evidence_links.owner_id IS '任务与证据共同所属的内部用户 UUID';
COMMENT ON COLUMN task_evidence_links.task_id IS '正式任务 UUID';
COMMENT ON COLUMN task_evidence_links.source_evidence_id IS '来源证据 UUID';
COMMENT ON COLUMN task_evidence_links.created_at IS '关联创建时间';

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
  CONSTRAINT chk_pending_target_type CHECK (char_length(target_type) > 0),
  CONSTRAINT chk_pending_action CHECK (char_length(action) > 0),
  CONSTRAINT chk_pending_parameters_object CHECK (jsonb_typeof(parameters) = 'object'),
  CONSTRAINT chk_status CHECK (status IN ('pending', 'executed', 'cancelled', 'expired', 'failed')),
  CONSTRAINT chk_expires_at CHECK (expires_at > created_at)
);
COMMENT ON TABLE pending_actions IS '等待用户再次确认的操作建议';
COMMENT ON COLUMN pending_actions.id IS '待确认操作 UUID';
COMMENT ON COLUMN pending_actions.owner_id IS '所属内部用户 UUID';
COMMENT ON COLUMN pending_actions.target_type IS '待操作目标的实体类别';
COMMENT ON COLUMN pending_actions.target_id IS '待操作目标 UUID';
COMMENT ON COLUMN pending_actions.target_version IS '提议时的目标版本';
COMMENT ON COLUMN pending_actions.action IS '拟执行的业务动作';
COMMENT ON COLUMN pending_actions.parameters IS '动作参数 JSON 对象，默认空对象';
COMMENT ON COLUMN pending_actions.status IS '确认状态：pending、executed、cancelled、expired 或 failed';
COMMENT ON COLUMN pending_actions.expires_at IS '确认期限';
COMMENT ON COLUMN pending_actions.created_at IS '操作提议时间';
COMMENT ON COLUMN pending_actions.updated_at IS '操作状态最后修改时间';

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
  CONSTRAINT chk_audit_metadata_object CHECK (jsonb_typeof(metadata) = 'object'),
  CONSTRAINT chk_audit_actor_type CHECK (char_length(actor_type) > 0),
  CONSTRAINT chk_audit_actor_id CHECK (char_length(actor_id) > 0),
  CONSTRAINT chk_audit_action CHECK (char_length(action) > 0),
  CONSTRAINT chk_audit_entity_type CHECK (char_length(entity_type) > 0),
  CONSTRAINT chk_audit_entity_id CHECK (char_length(entity_id) > 0)
);
COMMENT ON TABLE audit_events IS '只追加的业务操作审计记录';
COMMENT ON COLUMN audit_events.id IS '审计事件 UUID';
COMMENT ON COLUMN audit_events.owner_id IS '所属内部用户 UUID';
COMMENT ON COLUMN audit_events.actor_type IS '操作者类别';
COMMENT ON COLUMN audit_events.actor_id IS '操作者标识';
COMMENT ON COLUMN audit_events.action IS '发生的业务动作';
COMMENT ON COLUMN audit_events.entity_type IS '被操作实体的类别';
COMMENT ON COLUMN audit_events.entity_id IS '被操作实体的标识';
COMMENT ON COLUMN audit_events.metadata IS '允许记录的操作摘要 JSON 对象';
COMMENT ON COLUMN audit_events.request_id IS '关联 HTTP 请求的文本 ID，后台操作可为空';
COMMENT ON COLUMN audit_events.created_at IS '操作发生时间';
