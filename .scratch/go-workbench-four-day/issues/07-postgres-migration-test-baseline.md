# 07 — 启动隔离 PostgreSQL，用 SQL 建表并测试约束

**前置条件：** 06 验收通过后才能执行本任务。文档准备不表示前置任务已通过。

**状态：** Agent 配置、runner 和测试空壳已交付，业务 SQL 和业务断言待你填写。`migrations/` 当前只含说明文件，没有可执行的业务 migration；完整 07 仍未验收。runner 工具测试与业务测试分开执行，不能用工具测试通过代替业务验收。

## 到底要干什么

你要在专门用于测试的 PostgreSQL 数据库中，用 SQL 创建业务表，并用 Go 测试证明数据库会拒绝错误数据。

例如，05 的 `LocalRepository` 返回写在 Go 代码中的示例任务。07 先准备真正存放任务的 `tasks` 表：它包含任务 ID、所属用户、标题、状态等列。你要规定任务必须属于存在的用户、状态只能使用允许值、版本必须大于零。测试会尝试插入 `status = 'unknown'` 的任务，确认数据库拒绝它，而且没有留下非法记录。

07 交付的是可以存数据且有约束的数据库。08 再写 PostgreSQL Repository，让 Task API 查询这些表。07 不修改首页，不新增 HTTP API，也不替换 `LocalRepository`。

## 为什么要做

API、worker 和 MCP 会访问同一个数据库。只在某一段 Go 代码中校验数据，无法保护其他写入路径。因此关键规则还需要数据库约束，并用真实 PostgreSQL 测试验证。

建表 SQL 必须保存在仓库中。别人拿到空测试库后，可以执行同一份文件得到相同结构；后续增加列或约束，也可以用新文件记录变化。

## 先分清你要写什么

| 名称 | 作用 | 本任务写什么 |
| --- | --- | --- |
| PostgreSQL | 保存数据、执行 SQL 的数据库进程 | Agent 用 Compose 配置本地测试服务 |
| migration | 按版本保存的数据库结构变更 SQL | 你写 `CREATE TABLE`、列、约束和必要索引 |
| migration runner | 连接数据库并执行尚未应用的 SQL 文件的 Go 程序 | Agent 提供启动入口、事务和版本记录 |
| PostgreSQL Repository | 用 SQL 读取或修改业务数据的 Go 实现 | 属于 08，本任务不实现 |

你的主要实现文件是 `migrations/000001_initial.up.sql`。它是 SQL 文件，不是 Go 结构体或接口。你还要填写数据库测试的操作和断言。

## 文件和分工

以下路径作为本任务的执行约定。若实现时调整入口，必须同步更新文档命令。

| 文件 | 谁负责 | 必须写什么 |
| --- | --- | --- |
| `deploy/compose.yaml` | Agent | 本地 PostgreSQL、固定版本、healthcheck、回环地址端口、独立测试数据卷 |
| `deploy/test.env.example` | Agent | 非敏感配置示例和占位值，不保存真实凭据 |
| `cmd/migrate/main.go` | Agent | 调用 runner；失败时非零退出，不输出连接字符串 |
| `internal/postgres/migrate.go` | Agent | 检查连接目标、读取版本 SQL、事务执行、记录 migration 历史 |
| `internal/postgres/test_helpers_test.go` | Agent | 隔离测试 schema、测试数据和清理辅助函数 |
| `internal/postgres/migrate_test.go` | Agent 建空壳，你补测试 | 空库建表、重复执行、执行失败回滚、防误连 |
| `internal/postgres/schema_test.go` | Agent 建空壳，你补测试 | 关键列和约束、非法写入的失败断言 |
| `internal/postgres/runner_test.go` | Agent | 已实现的工具测试：文件排序、连接保护、重复执行、失败回滚、校验值变更、并发锁 |
| `migrations/000001_initial.up.sql` | 你 | 第一版完整业务建表 SQL |

Agent 可以引入 `github.com/jackc/pgx/v5`：仓库当前没有 PostgreSQL 驱动，Go 标准库不能独立建立 PostgreSQL 协议连接。你不需要先实现 runner，才能开始学习建表 SQL。

## 第 1 步：先准备可运行的工具

让 Agent 交付上表中负责的配置、runner 和测试空壳，暂不代写你的业务建表 SQL。Compose 服务名用 `db`，数据库名用 `go_workbench_test`，宿主机端口用 `127.0.0.1:5433`。端口占用时修改端口，并同步连接目标检查和配置示例。

runner 和测试通过 `TEST_DATABASE_URL` 获取连接配置，由你在本地安全设置。配置格式见 `deploy/test.env.example`；Compose 使用 `TEST_POSTGRES_PASSWORD`，URL 中的密码必须与它一致，并进行 URL 编码。连接目标固定为 `127.0.0.1:5433/go_workbench_test`，用户名为 `workbench_test`，查询参数只允许 `sslmode=disable`。Agent 不读取真实环境文件，不打印变量值，命令中不写真实连接字符串。

Compose 和 Go 命令都读取当前终端的环境变量；Go 不会自动加载配置文件。下面的 Compose 命令显式使用 `/dev/null`，避免自动加载仓库中的真实 `.env`。已有数据卷再次启动时必须沿用原密码，修改环境变量不会重置数据库中的密码。

runner 和测试都必须先检查配置存在、目标为专用本地测试实例、数据库名恰好为 `go_workbench_test`；连接后再核对实际数据库名。检查失败立即停止，不能退回其他数据库。防误连测试使用虚构配置，不实际连接非测试数据库。

```bash
docker compose --env-file /dev/null -f deploy/compose.yaml config --quiet
```

**作用：** 检查 Compose 能否解析；`--quiet` 避免把解析后的凭据输出到终端。

**通过标准：** 退出码为 0。配置缺失应先处理，不能当作 migration 的预期失败。

## 第 2 步：启动专用测试数据库

```bash
docker compose --env-file /dev/null -f deploy/compose.yaml up -d --wait db
docker compose --env-file /dev/null -f deploy/compose.yaml ps db
```

`up -d db` 在后台启动数据库，`ps db` 查看状态。等到 `healthy` 后再执行 SQL。只有 `running` 不等于数据库已经可以连接。

失败时在本地检查 `docker compose --env-file /dev/null -f deploy/compose.yaml logs db`，分享前移除敏感信息。不要通过删除数据卷或 `down -v` 试错。

**通过标准：** 数据库健康，runner 可以连接经过核对的专用测试库。Docker 不可用、端口占用和连接失败属于环境问题，不是预期的 RED。

## 第 3 步：逐张写出建表 SQL

对照 [总计划的数据模型](../../../GO_WORKBENCH_4_DAY_PLAN.md)、[业务术语](../../../CONTEXT.md) 和现有 `internal/task/service.go`，在 `migrations/000001_initial.up.sql` 中写十张业务表。以下是**待你确认并实现的第一版字段方案**，不是已执行的 schema，也不是已有产品契约。标记“可空”的列允许 `NULL`；其余列建议 `NOT NULL`。`id` 建议统一用 `text`，与当前 Go `Task.ID string` 兼容；由写入方生成且不能为空。时间统一用 `timestamptz`，创建时间默认 `now()`。字段名、ID 类型、存储范围和长度上限是当前总计划未定的选择；如果你更改它们，要同步调整测试 fixture 和后续 08–17 的 SQL。不要为尚未确定的来源类型或动作类型编造枚举。

### 1. `owners`：谁拥有工作台数据

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `id` | `text` | 内部 Owner ID，主键；其他表用它表示归属 |
| `oidc_issuer` | `text` | 登录令牌的签发方；不得为空字符串 |
| `oidc_subject` | `text` | 该签发方下的用户 ID；不得为空字符串 |
| `created_at` | `timestamptz` | 内部 Owner 创建时间 |

给 `(oidc_issuer, oidc_subject)` 加 `UNIQUE`：同一外部身份只对应一个内部 Owner。07 只建映射字段，17 才验证令牌并填入可信身份；不从客户端直接读取 `owner_id`。

### 2. `projects`：Owner 创建的项目

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `id` | `text` | Project ID，主键 |
| `owner_id` | `text` | 所属 Owner；外键指向 `owners(id)` |
| `name` | `text` | 项目名称；不得为空字符串 |
| `description` | `text`，可空 | 项目说明 |
| `status` | `text` | 仅 `active`、`archived`；默认 `active` |
| `created_at` | `timestamptz` | 创建时间 |
| `updated_at` | `timestamptz` | 最后修改时间 |

Task 需要按 Owner 关联 Project，因此除了主键，还要支持 `(owner_id, id)` 复合外键。Project 归档不自动归档现有 Task；禁止新任务分配给已归档 Project 需要后续写事务检查，不能仅靠这里的状态 `CHECK` 完成。

### 3. `tasks`：已明确创建或接受的正式任务

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `id` | `text` | Task ID，主键；映射 `task.Task.ID` |
| `owner_id` | `text` | 所属 Owner；外键指向 `owners(id)`，映射 `OwnerID` |
| `project_id` | `text`，可空 | 所属 Project；与 `owner_id` 一起关联同一 Owner 的 Project；直接创建的 Task 可以不选项目 |
| `title` | `text` | 标题；不得为空字符串，映射 `Title` |
| `description` | `text`，可空 | 任务详情 |
| `status` | `text` | 进度，限 `backlog`、`in_progress`、`blocked`、`done`；映射 `Status` |
| `priority` | `text` | 优先级，限 `none`、`low`、`medium`、`high`；映射 `Priority` |
| `labels` | `text[]` | 标签；默认空数组而非 `NULL`，映射 `Labels`；Label 不单独建表 |
| `due_at` | `timestamptz`，可空 | 可选截止时间 |
| `archived_at` | `timestamptz`，可空 | 归档时间；`NULL` 表示未归档，不作为 `status` 值 |
| `version` | `bigint` | 乐观并发版本；初值 `1`，`CHECK (version > 0)` |
| `created_at` | `timestamptz` | 创建时间，映射 `CreatedAt` |
| `updated_at` | `timestamptz` | 修改时间，映射 `UpdatedAt` |

直接创建的 Task 可以没有来源证据；由 Candidate 接受的 Task 可关联多条证据。08 负责把 SQL 列映射为现有 Go `task.Task`，此处只定义存储结构。

### 4. `check_items`：一个 Task 内的步骤

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `id` | `text` | 检查项 ID，主键 |
| `task_id` | `text` | 所属 Task；外键指向 `tasks(id)` |
| `content` | `text` | 步骤内容；不得为空字符串 |
| `completed` | `boolean` | 是否完成；当前 SQL 默认 `false`，但允许显式写入 `NULL` |
| `position` | `integer` | 在 Task 内的显示顺序；`CHECK (position >= 0)`；当前 SQL 还要求同一 Task 内唯一 |
| `created_at` | `timestamptz` | 创建时间；当前 SQL 默认 `now()`，但允许显式写入 `NULL` |
| `updated_at` | `timestamptz` | 修改时间；当前 SQL 默认 `now()`，但允许显式写入 `NULL` |

这是 Task 的步骤，不是独立 Task；它通过 `task_id` 继承 Task 归属，不需要单独的 Project 或优先级。当前 SQL 用 `UNIQUE (task_id, position)` 防止同一 Task 内的位置重复；排序时如何调整位置仍需在编辑行为中确定。`DEFAULT` 只在写入方省略该列时生效；若业务不允许 `completed` 或时间列为 `NULL`，仍需加 `NOT NULL`。

### 5. `captures`：一次原始输入及其处理进度

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `id` | `text` | Capture ID，主键 |
| `owner_id` | `text` | 所属 Owner；外键指向 `owners(id)` |
| `idempotency_key` | `text` | 提交方提供的精确重试身份；与 `owner_id` 联合唯一，不能为空字符串 |
| `input_hash` | `text` | 输入内容摘要，用于判断同一键对应的输入是否变化；建议 SHA-256 十六进制值，长度 64 |
| `input_text` | `text` | Worker 处理所需的最小输入内容；写入前限制长度，不保存完整对话档案 |
| `source_type` | `text` | 来源类别；暂不枚举未确认的类型 |
| `status` | `text` | 限 `queued`、`processing`、`processed`、`failed`；默认 `queued` |
| `attempt_count` | `integer` | 处理尝试次数；默认 `0`，不得为负 |
| `last_error` | `text`，可空 | 脱敏且有长度上限的失败摘要，不保存原始异常或凭据 |
| `created_at` | `timestamptz` | 捕获时间 |
| `updated_at` | `timestamptz` | 状态更新时间 |

Capture 是输入记录，不是 Task。`UNIQUE (owner_id, idempotency_key)` 防止同一 Owner 重复创建；09 负责比对 `input_hash` 并返回原 Capture 或 `409`。原始输入不可改写；状态、尝试次数和失败摘要需要 Worker 更新，不能把整个 Capture 表设为完全不可更新。

### 6. `candidates`：从 Capture 生成、等待人工审核的建议

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `id` | `text` | Candidate ID，主键 |
| `owner_id` | `text` | 所属 Owner；与 `capture_id` 一起证明来源属于同一 Owner |
| `capture_id` | `text` | 唯一来源 Capture；外键关联同 Owner 的 `captures` |
| `proposed_title` | `text` | 建议标题；不得为空字符串，用户接受前可修改 |
| `proposed_description` | `text`，可空 | 建议详情；接受前可修改 |
| `proposed_project_id` | `text`，可空 | 建议 Project；若有值必须属于同一 Owner |
| `labels` | `text[]` | 建议标签；默认空数组，接受后复制到 Task |
| `status` | `text` | 限 `pending_review`、`accepted`、`rejected`；默认 `pending_review` |
| `accepted_task_id` | `text`，可空 | 接受后创建的 Task；若有值必须属于同一 Owner，一个 Task 最多被一个 Candidate 认领 |
| `created_at` | `timestamptz` | 建议创建时间 |
| `updated_at` | `timestamptz` | 审核或编辑时间 |

一个 Capture 可产生零到多个 Candidate；一个 Candidate 只来自一个 Capture。Inbox 是查询 `pending_review` Candidate 的视图，不需要 `inbox` 表。10 的事务才负责“接受一次、创建 Task、关联证据、更新 Candidate”。

### 7. `source_evidence`：不可改写的来源引用

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `id` | `text` | Evidence ID，主键 |
| `owner_id` | `text` | 所属 Owner；与 `capture_id` 一起确保来源同 Owner |
| `capture_id` | `text` | 对应原始 Capture；外键关联同 Owner 的 `captures` |
| `source_type` | `text` | 来源类别；与 Capture 来源对应 |
| `external_ref` | `text`，可空 | 外部事件或对话的最小定位 ID；无外部 ID 时可空 |
| `source_url` | `text`，可空 | 可访问的来源位置，不保存访问凭据 |
| `excerpt` | `text`，可空 | 能解释任务来历的最小摘录；写入前控制长度和敏感内容 |
| `consent_scope` | `text`，可空 | 来自授权来源时的授权范围；手工录入时可空 |
| `created_at` | `timestamptz` | 证据保存时间 |

Evidence 只保存来源所需的最小信息，不能当作完整对话备份。来源引用写入后不可原地修改；只写 SQL 列和外键还不能证明不可变，后续写入权限或触发器需要单独验证。

### 8. `task_evidence_links`：正式任务与证据的多对多关联

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `owner_id` | `text` | 关联所属 Owner；参与两侧复合外键，防止跨 Owner 关联 |
| `task_id` | `text` | 被证据支持的 Task；与 `owner_id` 一起引用 `tasks` |
| `source_evidence_id` | `text` | 证据 ID；与 `owner_id` 一起引用 `source_evidence` |
| `created_at` | `timestamptz` | 建立关联的时间 |

建议用 `(task_id, source_evidence_id)` 作复合主键，避免同一对重复，同时保留 `owner_id` 进行归属校验。直接创建 Task 时这张表可以没有对应行；Candidate 接受事务负责写入需要的关联。

### 9. `pending_actions`：等待再次确认的动作

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `id` | `text` | Pending Action ID，主键 |
| `owner_id` | `text` | 所属 Owner；外键指向 `owners(id)` |
| `target_type` | `text` | 目标实体类别；目前主要是 Task |
| `target_id` | `text` | 待操作实体 ID |
| `target_version` | `bigint` | 提议时记录的目标版本；大于零，确认时重新比较 |
| `action` | `text` | 拟执行动作，如完成或归档；具体编码在 12 确认 |
| `parameters` | `jsonb` | 动作参数；默认空对象，只允许业务定义的字段和大小 |
| `status` | `text` | 限 `pending`、`executed`、`cancelled`、`expired`、`failed`；默认 `pending` |
| `expires_at` | `timestamptz` | 过期时间；不得早于创建时间 |
| `created_at` | `timestamptz` | 提议时间 |
| `updated_at` | `timestamptz` | 状态更新时间 |

`target_type` + `target_id` 是多种目标共用的引用，普通外键不能随 `target_type` 自动指向不同表；12 的确认事务要验证目标存在、归属、版本与过期时间。07 不因建了这些列就宣称确认规则已实现。

### 10. `audit_events`：只追加的操作记录

| 字段 | 建议类型 | 用途与约束 |
| --- | --- | --- |
| `id` | `text` | Audit Event ID，主键 |
| `owner_id` | `text` | 所属 Owner；外键指向 `owners(id)` |
| `actor_type` | `text` | 操作者类别，用来区分用户与系统进程 |
| `actor_id` | `text` | 操作者标识；不得为空字符串 |
| `action` | `text` | 发生的领域动作名称 |
| `entity_type` | `text` | 被操作的实体类别 |
| `entity_id` | `text` | 被操作的实体 ID |
| `metadata` | `jsonb` | 只存允许字段的摘要；默认空对象，不保存请求正文或凭据 |
| `request_id` | `text`，可空 | 有 HTTP 请求时用于关联日志；后台操作可空 |
| `created_at` | `timestamptz` | 操作发生时间 |

Audit Event 只追加；`entity_type` + `entity_id` 同样不能由普通外键动态指向不同表。10、12 等写事务需与对应业务变更原子写入事件。只定义表结构不能证明“只追加”或 metadata 安全，后续要限制写入路径并测试。

### Runner 自用表：`schema_migrations`

此表由已交付的 runner 自动创建，**不要**写进 `000001_initial.up.sql`。它包含 `version bigint`（主键）、`name text`（migration 文件名）、`checksum text`（文件内容 SHA-256）和 `applied_at timestamptz`（执行时间）。它用于判断哪些 SQL 已执行，以及已执行文件是否被改动；它不是第十一张业务表。

### 这些列之间必须满足什么关系

- 建表顺序建议：`owners` → `projects`、`tasks`、`captures`、`pending_actions`、`audit_events` → `check_items`、`candidates`、`source_evidence` → `task_evidence_links`。`candidates.accepted_task_id` 指向已创建的 `tasks`。
- 每个有 `owner_id` 的业务表都关联 `owners(id)`。跨表引用除了 ID 存在，还需同 Owner：在被引用表上支持 `(owner_id, id)` 唯一键，在引用表上使用 `(owner_id, 目标_id)` 复合外键。`tasks.project_id` 可空时，该外键允许无项目的 Task。
- `check_items` 通过 Task 继承归属，不单独存 `owner_id`；读取与修改检查项时仍必须通过 Task 验证 owner scope。`task_evidence_links` 同时关联同 Owner 的 Task 和 Evidence。
- PostgreSQL 外键拒绝不存在或跨 Owner 的记录，UNIQUE 拒绝重复，CHECK 限制单行状态与数值；“接受一次”“禁止分配到已归档 Project”“旧版本确认失败”“证据和审计不可改写”还需要后续事务、权限或触发器证明。

**需要你确认的设计选择：** ID 采用 `text` 还是 UUID；`input_text`、`excerpt`、`metadata` 的保存范围与上限；`source_type`、`action`、`actor_type` 的合法值；是否在 07 用触发器或权限强制不可变。这些没有在现有计划中定死。当前清单给出可执行起点，但不能把建议当作已确认的生产契约。

具体按以下顺序写：

1. 每张实体表先写主键和必需列，为必需列加 `NOT NULL`。它只拒绝 NULL，不拒绝空字符串；需要非空文本时另写 `CHECK`。
2. 创建 `owners`，再写 `projects` 和 `tasks`。字段要与 `internal/task/service.go` 的 `Task` 对齐；SQL 列名如 `owner_id`、`created_at`，Go 字段映射留到 08。
3. 为 Task 状态、优先级和 version 写约束。下面是列定义示例，你需要把它放入完整的 `CREATE TABLE tasks (...)` 中并补齐其他列：

   ```sql
   status text NOT NULL
       CHECK (status IN ('backlog', 'in_progress', 'blocked', 'done')),
   version bigint NOT NULL DEFAULT 1 CHECK (version > 0)
   ```

4. Task priority 使用现有 Go 类型的 `none | low | medium | high`；归档用独立 `archived_at`，不能加 `archived` 状态。Capture 状态用 `queued | processing | processed | failed`；Candidate 用 `pending_review | accepted | rejected`；Pending Action 用 `pending | executed | cancelled | expired | failed`。分别写 CHECK，不共用状态集合。
5. 为所属用户和引用关系写外键。Task → Project、Candidate → Capture、Evidence → Capture 等关联还必须保证同一 owner。可使用 `(owner_id, id)` 唯一约束与复合外键；双方各自有 owner 外键不能阻止跨用户关联。
6. 为 Capture 写 `UNIQUE (owner_id, idempotency_key)`。同一 owner 不能重复占用同一键，不同 owner 可以用同一键。相同键不同输入的 `409` 处理属于 09。
7. 为 `task_evidence_links` 写 Task ID 与 Evidence ID 的联合唯一约束。为 Candidate 接受后关联的 Task 明确唯一性，避免多个 Candidate 认领同一 Task；完整接受事务属于 10。
8. 补齐剩余表，为后续 owner 查询和关联查询加必要索引。外键不会自动给引用方创建索引，不要给每列都加索引。
9. 不配置级联删除原始输入、证据和审计记录。Project 归档不能级联修改 Task。禁止分配给已归档项目、证据不可变和状态转换是否合法，还需要后续事务或权限控制测试，不能仅靠枚举 CHECK 证明。

**通过标准：** 十张表都写入 SQL；你能说明每个约束防止什么错误。这里只证明 SQL 已写好，还没有证明它能执行。

## 第 4 步：先写 migration 测试

Agent 提供隔离辅助函数：每个测试创建唯一命名的测试 schema，使用明确指定 `search_path` 的连接执行 migration 和断言，结束后只清理自己创建的 schema。历史表也放在对应 schema 内。不要删除整个数据库，不依赖其他测试残留的数据。

你在 `internal/postgres/migrate_test.go` 填写前三项测试；第四项防误连工具测试已在 `runner_test.go` 实现。其他业务测试空壳都有明确的 `t.Fatal("RED: ...")`，填写完操作和断言后移除对应占位失败。不要直接删掉占位失败而不添加断言。

| 测试名称 | 具体操作 | 必须断言 |
| --- | --- | --- |
| `TestMigrationEmptySchema` | 对无业务表的测试 schema 执行 runner | 十张表、关键列和命名约束存在；历史表有版本 1 |
| `TestMigrationRepeat` | 应用一次，插入合法 Owner/Task，再应用一次 | 已应用版本跳过；历史不重复；既有数据保留 |
| `TestMigrationAtomicFailure` | 使用测试专用 SQL，建表后故意执行错误 SQL | 返回错误；没有部分建表结果，也没有成功版本记录 |
| `TestMigrationRejectUnsafeTarget` | 传入缺失或非测试目标的虚构配置 | 连接或执行 SQL 前拒绝；日志没有连接字符串 |

```bash
go test ./internal/postgres -run '^TestMigration' -count=1 -v
```

**RED 判据：** 连接和隔离正常，测试因目标表或约束缺失而失败。没有匹配测试、`no test files`、`SKIP`、数据库未启动都不算有效验证。

已交付的 runner 使用一个事务执行本次所有待应用版本及历史写入；任一版本失败，整次待应用批次回滚。事务级 advisory lock 串行化同一 schema 的并发迁移。SQL 文件不能包含自身的 `BEGIN`、`COMMIT`、`ROLLBACK`、显式业务 schema 名或 `SET search_path`，具体限制见 `migrations/README.md`。已应用文件修改、丢失或插入旧版本时会拒绝执行。业务 SQL 不要只依赖 `CREATE TABLE IF NOT EXISTS`，它不会验证已有表的列和约束是否正确。

业务 SQL 尚未编写时，可以单独验证 Agent 工具：

```bash
# 不需要数据库：文件读取、连接保护和错误脱敏。
go test ./internal/postgres -run '^(TestRunnerUnit|TestMigrationRejectUnsafeTarget)' -count=1 -v

# 需要已配置、健康的隔离测试数据库：重复执行、回滚、文件修改和并发锁。
go test -race ./internal/postgres -run '^TestRunnerIntegration' -count=1 -v
```

工具集成测试仅在独立测试 schema 中创建临时 probe 表，不使用你的业务 SQL。业务测试辅助函数已提供 `newTestDatabase`、`applyBusinessMigration` 和 `assertPgError`；Owner/Task fixture 的列和值由你按最终建表 SQL填写。

## 第 5 步：执行建表并重跑测试

runner 交付以下入口。从仓库根目录执行，连接配置只通过前面设置的环境变量读取：

```bash
go run ./cmd/migrate -dir migrations up
go run ./cmd/migrate -dir migrations up
go test ./internal/postgres -run '^TestMigration' -count=1 -v
```

第一次创建表并记录版本。第二次报告没有待应用的版本，不重复建表、不清空数据。SQL 错误或测试失败时先修复再验证。

**通过标准：** 两次 runner 退出码为 0，目标测试实际运行并通过。除了检查输出，还要由测试查询实际表结构和数据。

## 第 6 步：测试数据库会拒绝哪些错误

在 `internal/postgres/schema_test.go` 中写测试。每例先应用 migration，插入所需合法父记录，再执行一次故意违规的 INSERT。否则失败可能来自无关必填字段，不能证明目标约束有效。

| 测试名称 | 故意写入什么 | 预期结果 |
| --- | --- | --- |
| `TestSchemaInvalidTaskStatus` | Task 状态为 `unknown` | CHECK 拒绝，SQLSTATE `23514` |
| `TestSchemaDuplicateCaptureKey` | 同一 owner 两条相同幂等键 Capture | 唯一约束拒绝，SQLSTATE `23505` |
| `TestSchemaCaptureKeyAcrossOwners` | 两个 owner 各用一次相同幂等键 | 两条记录都能插入 |
| `TestSchemaMissingOwner` | Task 引用不存在的 owner | 外键拒绝，SQLSTATE `23503` |
| `TestSchemaCrossOwnerProject` | Owner A 的 Task 引用 Owner B 的 Project | owner 复合外键拒绝，SQLSTATE `23503` |
| `TestSchemaDuplicateEvidenceLink` | 重复插入同一 Task/Evidence 关联 | 唯一约束拒绝，SQLSTATE `23505` |

用 `errors.As` 取得 `*pgconn.PgError`，检查 SQLSTATE 和你在 SQL 中命名的约束。只检查 `err != nil` 不够，拼错 SQL 也会产生错误。为其他状态、引用、version 和尝试次数约束补对应测试。

事务中一条语句失败后，PostgreSQL 会使该事务进入失败状态。先回滚失败事务或回到 SAVEPOINT，再查询记录数：非法行必须为零，之前合法的行必须保留。不要直接在失败事务里继续 SELECT。

```bash
go test ./internal/postgres -run '^TestSchema' -count=1 -v
go test ./... -count=1
git diff --check
```

**通过标准：** 目标测试实际执行并通过，完整 Go 测试通过，diff 无格式问题。缺少测试配置必须明确报错，不能静默跳过后声称通过。

## 第 7 步：记录升级和回滚限制

已应用 migration 不再修改。以后新增列或约束时增加 `000002_*.up.sql`，并测试从版本 1 升级且保留旧数据。现在只有第一版，“重复执行版本 1”只能证明重复执行安全，不能声称未来版本升级兼容性已通过。

本阶段不提供自动 `down`，不执行生产 migration。建表过程中失败由事务回滚；成功建表后删除表会丢失数据，必须单独评估、备份并获得确认。删除数据卷不是回滚方案。

## 最终验收和证据

### 本次 Agent 交付验证

- Compose 使用示例占位配置执行 `config --quiet`，退出码 0；未读取真实配置。
- 全部 Go 包编译通过；`go vet ./cmd/migrate ./internal/postgres` 通过。
- 已有 `cmd/api`、`internal/httpapi`、`internal/task` 测试通过。
- 使用固定摘要 PostgreSQL 16 的一次性本地容器，执行 `go test -race ./internal/postgres -run '^(TestRunner|TestMigrationRejectUnsafeTarget)' -count=1 -v`，全部工具测试通过，包括真实事务回滚、重复执行、校验值变更拒绝及并发锁。容器已停止，内存测试数据已清理，没有改动现有数据卷。
- 明确移除 `TEST_DATABASE_URL` 后运行完整测试，数据库测试因缺少配置失败。这验证了不静默跳过的行为，不是有效业务 RED，也不是完整测试通过。
- 业务 SQL 和业务断言尚未填写；十张表、业务约束和完整 07 验收均未通过。下一步先设置本地环境变量、启动 Compose，再按第 3 步编写 SQL。

- [ ] 06 已验收；测试数据库健康；连接目标经过核对。
- [ ] 十张业务表与总计划一致，包括原文遗漏的 `check_items`。
- [ ] 空 schema、重复执行、失败回滚和防误连测试实际通过。
- [ ] CHECK、UNIQUE、外键和跨 owner 关联都有真实数据库失败测试。
- [ ] 不同 owner 可以复用幂等键；非法写入不影响已有合法数据。
- [ ] migration 历史、已应用文件修改检测和并发串行保护已验证。
- [ ] 测试只清理自己创建的 schema，没有删除数据库或数据卷。
- [ ] 交付 SQL、测试代码和脱敏后的实际命令结果；未执行、SKIP、环境失败分别记录。
- [ ] 回滚限制已记录；后续版本升级兼容性标记为待新增 migration 后验证。

完成后解释：建表 SQL、runner 和 Repository 各负责什么；CHECK、UNIQUE、外键分别防止什么错误；为什么普通单列外键不能保证关联数据属于同一 owner。验收通过后进入 08，实现从数据库读取 Task。
