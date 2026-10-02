# 07：用 PostgreSQL 建表并验证约束

**前置：** 06 验收通过后执行。07 只负责数据库结构和数据库测试；08 才实现 PostgreSQL Repository，不在这里替换 `LocalRepository`。

**当前状态：** [建表 SQL](../../../migrations/000001_initial.up.sql)、Compose、迁移执行器和测试空壳已存在。业务测试仍有 `RED` 占位；本文件不声称建表 SQL 或业务约束已经通过真实数据库测试。

## 要完成什么

1. 在专用测试库中执行 `000001_initial.up.sql`，得到十张业务表。
2. 测试首次执行、重复执行、失败回滚。
3. 插入合法与非法数据，验证 `CHECK`、`UNIQUE` 和外键。

API、Worker 和 MCP 都可能写入数据库。数据库约束保护共同的数据规则；跨表状态转换仍由后续业务事务处理。

## 文件位置

| 文件 | 作用 |
| --- | --- |
| [`migrations/000001_initial.up.sql`](../../../migrations/000001_initial.up.sql) | 你写的十张业务表、字段和约束 |
| [`deploy/compose.yaml`](../../../deploy/compose.yaml) | 专用 PostgreSQL 测试服务 |
| [`deploy/test.env.example`](../../../deploy/test.env.example) | 非敏感配置格式示例 |
| [`internal/postgres/migrate.go`](../../../internal/postgres/migrate.go) | 迁移执行器：事务、版本记录、防误连 |
| [`internal/postgres/migrate_test.go`](../../../internal/postgres/migrate_test.go) | 三个待补断言的迁移测试 |
| [`internal/postgres/schema_test.go`](../../../internal/postgres/schema_test.go) | 六个待补断言的约束测试 |
| [`internal/postgres/test_helpers_test.go`](../../../internal/postgres/test_helpers_test.go) | 每个测试独立建 schema，结束后清理 |
| [`internal/postgres/runner_test.go`](../../../internal/postgres/runner_test.go) | 已写好的执行器工具测试 |

## 第 1 步：准备本地配置

按 [`deploy/test.env.example`](../../../deploy/test.env.example) 在本地设置 `TEST_POSTGRES_PASSWORD` 和 `TEST_DATABASE_URL`。URL 中的密码必须与 Compose 密码一致，并按 URL 规则编码。Go 不会自动读取配置文件。不要把真实值写进代码、文档或命令输出。测试只接受 `127.0.0.1:5433/go_workbench_test` 和用户 `workbench_test`；端口改变时，配置与代码中的目标检查也要一起改。

## 第 2 步：启动专用测试库

从仓库根目录运行：

```bash
docker compose --env-file /dev/null -f deploy/compose.yaml config --quie
docker compose --env-file /dev/null -f deploy/compose.yaml up -d --wait db
docker compose --env-file /dev/null -f deploy/compose.yaml ps db
```

`/dev/null` 让 Compose 不加载仓库里的真实 `.env`。`config --quiet` 只检查配置，不打印解析后的配置。`ps db` 应显示数据库健康；连接失败属于环境问题，不算业务测试的预期失败。不要用 `down -v` 清理数据卷。

## 第 3 步：确认建表范围

当前 SQL 使用 `uuid` ID，默认由数据库生成。下表按业务流排列字段；准确类型、默认值和约束名以 [SQL 文件](../../../migrations/000001_initial.up.sql)为准。

| 表 | 存什么；关键字段 |
| --- | --- |
| `owners` | 内部用户：`id`、`oidc_issuer`、`oidc_subject`、`created_at`。外部身份二元组唯一。 |
| `projects` | 项目：`id`、`owner_id`、`name`、`description`、`status`、`version`、`created_at`、`updated_at`。`status` 为 `active` 或 `archived`。 |
| `tasks` | 正式任务：`id`、`owner_id`、`project_id`、`title`、`description`、`status`、`priority`、`labels`、`due_at`、`archived_at`、`version`、`created_at`、`updated_at`。`project_id` 可空，非空时必须属于同一 Owner。 |
| `check_items` | 任务步骤：`id`、`task_id`、`content`、`completed`、`position`、`created_at`、`updated_at`。同一 Task 的位置唯一。 |
| `captures` | 一次输入：`id`、`owner_id`、`idempotency_key`、`input_hash`、`input_text`、`source_type`、`status`、`attempt_count`、`last_error`、`created_at`、`updated_at`。同一 Owner 的幂等键唯一。当前 SQL 将文本限制为 1 至 250 个字符；这仍需产品确认。 |
| `source_evidence` | 输入来源：`id`、`owner_id`、`capture_id`、`source_type`、`external_ref`、`source_url`、`excerpt`、`consent_scope`、`created_at`。Capture 必须属于同一 Owner。 |
| `candidates` | 待审核建议：`id`、`owner_id`、`capture_id`、`proposed_title`、`proposed_description`、`proposed_project_id`、`labels`、`status`、`accepted_task_id`、`created_at`、`updated_at`。接受后关联同一 Owner 的 Task。 |
| `task_evidence_links` | Task 与证据的关联：`owner_id`、`task_id`、`source_evidence_id`、`created_at`。同一对不可重复，两侧属于同一 Owner。 |
| `pending_actions` | 待再次确认的操作：`id`、`owner_id`、`target_type`、`target_id`、`target_version`、`action`、`parameters`、`status`、`expires_at`、`created_at`、`updated_at`。 |
| `audit_events` | 操作记录：`id`、`owner_id`、`actor_type`、`actor_id`、`action`、`entity_type`、`entity_id`、`metadata`、`request_id`、`created_at`。`request_id` 可空。 |

`schema_migrations` 是执行器创建的版本历史表，不是第十一张业务表。`check_items` 通过 Task 继承 Owner；`pending_actions.target_type + target_id` 不能靠普通外键动态指向不同表。

**边界：** 07 的外键、唯一和检查约束能拒绝部分错误数据。“不能分配到已归档项目”“确认时重新比较版本”“来源和审计记录不可改写”等规则还要在后续事务、权限或触发器中实现并测试。不要把建表成功当成这些规则已通过。

## 第 4 步：写迁移测试

打开 [`internal/postgres/migrate_test.go`](../../../internal/postgres/migrate_test.go)。三个测试已有函数名和 `RED` 占位。`newTestDatabase(t)` 会在专用库中创建本测试独有的空 schema；`applyBusinessMigration` 在该 schema 执行正式 SQL。你只补操作和断言。完成一个测试后，再删除该测试末尾的 `t.Fatal("RED: ...")`。

### 4.1 首次建表：`TestMigrationEmptySchema`

**目的：** 证明 SQL 建出十张业务表、关键列、关键约束，且历史表只记录版本 `1`。

在现有 `applyBusinessMigration` 后，先按表名查询。下面的查询对 `tasks` 返回 `true`；把表名放进循环，检查第 3 步列出的十张表，再检查 `schema_migrations`。

```go
var exists bool
err := conn.QueryRow(ctx, `SELECT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = $1 AND table_name = $2
)`, schema, "tasks").Scan(&exists)
if err != nil {
    t.Fatal(err)
}
if !exists {
    t.Fatal("tasks table is missing")
}
```

再查关键列：`information_schema.columns` 可读 `data_type`、`is_nullable`、`column_default`。至少确认 `tasks.id` 为 `uuid`、`tasks.version` 为非空 `bigint`，以及 `captures.input_hash` 为非空 `text`。查约束时按 schema、表名和约束名过滤 `pg_constraint`；至少确认 `fk_task_project`、`fk_candidate_capture`、`uk_capture_owner_id_idempotency_key`。把下面的 SQL 放进 `QueryRow(ctx, sql, 参数...).Scan(...)`，分别核对返回值：

```sql
-- 传入 schema、"tasks"、"version"；期望 bigint、NO、默认值 1。
SELECT data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_schema = $1 AND table_name = $2 AND column_name = $3;

-- 传入 schema、"tasks"、"fk_task_project"；期望 true。
SELECT EXISTS (
    SELECT 1 FROM pg_constraint AS c
    JOIN pg_class AS t ON t.oid = c.conrelid
    JOIN pg_namespace AS n ON n.oid = t.relnamespace
    WHERE n.nspname = $1 AND t.relname = $2 AND c.conname = $3
);
```

最后执行 `SELECT count(*), coalesce(min(version), 0) FROM schema_migrations`，用两个整数接收结果，期望 `1, 1`。单纯检查 `Migrate` 没报错，不足以证明表结构正确。

### 4.2 重复执行：`TestMigrationRepeat`

**目的：** 第二次执行不重建表，不删除数据，也不重复记录版本。

首次执行后插入一个虚构 Owner 和一条 Task。数据库生成 UUID，`RETURNING id::text` 让测试保存 ID。当前 Task 的 `description` 可空，因此最小 Task 只需 `owner_id` 和 `title`。

```go
var ownerID, taskID string
err := conn.QueryRow(ctx,
    "INSERT INTO owners (oidc_issuer, oidc_subject) VALUES ($1, $2) RETURNING id::text",
    "https://issuer.example.invalid", "test-subject",
).Scan(&ownerID)
if err != nil { t.Fatal(err) }

err = conn.QueryRow(ctx,
    "INSERT INTO tasks (owner_id, title) VALUES ($1::uuid, $2) RETURNING id::text",
    ownerID, "test task",
).Scan(&taskID)
if err != nil { t.Fatal(err) }

applied, err := Migrate(ctx, conn, "../../migrations", schema)
if err != nil { t.Fatal(err) }
if applied != 0 { t.Fatalf("applied=%d, want 0", applied) }
```

再查 `schema_migrations`，期望仍只有版本 `1`。按 `taskID` 查询 Task，确认标题和 `ownerID` 没变。`applied` 表示本次新执行的 migration 文件数，不是表数。

### 4.3 失败回滚：`TestMigrationAtomicFailure`

**目的：** SQL 中途失败时，执行器撤销本次建表和历史记录。

此测试从 `ctx, conn, schema := newTestDatabase(t)` 开始，不调用 `applyBusinessMigration`。复用 [`writeRunnerSQL`](../../../internal/postgres/runner_test.go) 写一份**临时** SQL；不要改正式建表文件。

```go
dir := t.TempDir()
writeRunnerSQL(t, dir, "000001_broken.up.sql",
    "CREATE TABLE rollback_probe (id integer); SELECT * FROM missing_probe_table;")
_, err := Migrate(ctx, conn, dir, schema)
if err == nil { t.Fatal("broken migration succeeded") }
```

然后分别执行 `SELECT to_regclass('rollback_probe') IS NULL` 和 `SELECT to_regclass('schema_migrations') IS NULL`。两次查询都应返回 `true`。首次迁移失败后历史表本身也应不存在，不能对它执行 `SELECT count(*)`。已有 [`TestRunnerIntegration`](../../../internal/postgres/runner_test.go)展示了同类查询。

### 4.4 运行并判断结果

从仓库根目录依次运行。每完成一个测试，再运行下一条：

```bash
go test ./internal/postgres -run '^TestMigrationEmptySchema$' -count=1 -v
go test ./internal/postgres -run '^TestMigrationRepeat$' -count=1 -v
go test ./internal/postgres -run '^TestMigrationAtomicFailure$' -count=1 -v
go test ./internal/postgres -run '^TestMigration' -count=1 -v
```

`RED: fill ...` 表示测试还没写完；`TEST_DATABASE_URL is required` 或连接失败表示环境未就绪；`apply migration ... failed` 表示 SQL 没执行成功。只有断言实际执行并全部通过，才算完成。`TestMigrationRejectUnsafeTarget` 已在 [`runner_test.go`](../../../internal/postgres/runner_test.go)实现；不要重复写。

## 第 5 步：验证数据库约束

打开 [`internal/postgres/schema_test.go`](../../../internal/postgres/schema_test.go)。每个测试先调用 `newTestDatabase` 和 `applyBusinessMigration`，再插入**合法的父记录**，最后只让一个字段或关联违规。这样测试失败才能归因于目标约束。

例如，先插入 Owner，再写状态为 `unknown` 的 Task。现成的 `assertPgError` 会同时核对 SQLSTATE 和约束名：

```go
_, err := conn.Exec(ctx,
    "INSERT INTO tasks (owner_id, title, status) VALUES ($1::uuid, $2, $3)",
    ownerID, "invalid status task", "unknown",
)
assertPgError(t, err, "23514", "chk_task_status")
```

这里的 `ownerID` 必须来自本测试刚插入的合法 Owner。随后查询，确认非法 Task 没写入。不要只断言 `err != nil`；SQL 拼写错误也会产生错误。

| 测试 | 违规输入 | 期望 |
| --- | --- | --- |
| `TestSchemaInvalidTaskStatus` | Task 状态 `unknown` | `23514`，`chk_task_status` |
| `TestSchemaDuplicateCaptureKey` | 同一 Owner 重复幂等键 | `23505`，`uk_capture_owner_id_idempotency_key` |
| `TestSchemaCaptureKeyAcrossOwners` | 不同 Owner 使用同一键 | 两次写入都成功 |
| `TestSchemaMissingOwner` | Task 引用不存在的 Owner | `23503`，Owner 外键 |
| `TestSchemaCrossOwnerProject` | A 的 Task 引用 B 的 Project | `23503`，`fk_task_project` |
| `TestSchemaDuplicateEvidenceLink` | 同一 Task/Evidence 关联两次 | `23505`，`uk_task_evidence_link_task_id_source_evidence_id` |

写 Capture 时还需满足 `input_hash` 为 64 位小写十六进制、`input_text` 为 1 至 250 字符，并提供非空 `source_type`；否则测试可能先撞上无关约束。某条语句在显式事务内失败后，先回滚事务或回到 savepoint，再查询行数；在普通连接上单独执行失败语句后，可以直接查询。

```bash
go test ./internal/postgres -run '^TestSchema' -count=1 -v
```

## 第 6 步：应用到测试库 public

第 4、5 步的测试都使用独立 schema，不会在 `public` 建业务表。确认测试库目标与现有数据后，再决定是否把 migration 应用到该**专用测试库的 `public` schema**。这一步会持久改变该 schema；如已有需要保留的数据，先备份并取得确认。

```bash
go run ./cmd/migrate -dir migrations up
go run ./cmd/migrate -dir migrations up
```

若 `public` 尚未执行版本 1，首次输出应为 `migrations applied: 1`，第二次应为 `migrations applied: 0`。若版本 1 已执行，执行器会核对文件校验值并返回 0；文件被改动则报错，不能继续修改已应用版本。执行器在同一事务中应用 SQL 和写入 `schema_migrations`。正式 migration 文件不能自行写 `BEGIN`、`COMMIT`、`ROLLBACK` 或 `SET search_path`。

## 第 7 步：总验收

最终运行：

```bash
go test ./internal/postgres -run '^TestMigration' -count=1 -v
go test ./internal/postgres -run '^TestSchema' -count=1 -v
go test ./... -count=1
git diff --check
```

**通过标准：** 十张表和历史版本 1 均存在；迁移重跑保留数据；失败迁移不留部分表；六个约束测试实际执行并通过；完整 Go 测试与 diff 检查通过。环境失败、`SKIP` 和仍有 `RED` 占位都不能算业务通过。

已应用的 migration 不再修改。以后加列或改约束时，新建 `000002_*.up.sql`。本阶段没有自动 `down`；不要用删除表或数据卷代替回滚。07 不证明未来版本升级、生产库迁移、业务状态转换或审计记录不可变。
