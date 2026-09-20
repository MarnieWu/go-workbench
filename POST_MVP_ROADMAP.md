# MVP 后路线图 / Post-MVP Roadmap

本文件防止 MVP 之后遗忘已经识别的工作，但不把所有可能功能承诺为必做项。每项只有在启动条件成立时进入开发。 / This document prevents known post-MVP work from being forgotten without treating every possible feature as a commitment. An item enters development only when its trigger is satisfied.

状态 / Status:

- `planned`：MVP 验收后立即评估并安排。 / Evaluate and schedule immediately after MVP acceptance.
- `trigger-based`：只有可观察条件成立时启动。 / Start only when its observable trigger is met.
- `rejected`：当前明确不做，重新考虑需要新的证据。 / Explicitly excluded; reconsider only with new evidence.

## A. MVP 后立即补强 / Immediate Post-MVP Hardening

### 1. 数据保留与隐私删除 / Retention and Privacy Deletion

- 状态 / Status: `planned`
- 价值与风险 / Value and risk: 控制 Capture、Candidate 和 Source Evidence 长期积累，并提供可恢复、可审计的个人数据删除。 / Control long-term accumulation and provide recoverable, auditable personal-data deletion.
- 启动条件 / Trigger: MVP 核心流程通过验收。 / MVP core flow passes acceptance.
- 前置依赖 / Dependencies: 实体关系、审计边界、备份恢复策略稳定。 / Stable entity relationships, audit boundary, and backup/recovery policy.
- 最小范围 / Minimum scope: 可配置保留期限、删除预览、明确确认、异步清理、失败恢复和审计 tombstone；不通过 MCP 暴露删除。 / Configurable retention, deletion preview, explicit confirmation, asynchronous cleanup, failure recovery, and an audit tombstone; no MCP delete tool.
- 验收证据 / Evidence: 时间边界测试、级联影响测试、恢复演练和敏感数据扫描。 / Time-boundary tests, cascade-impact tests, recovery drill, and sensitive-data scan.

### 2. 生产身份与会话安全 / Production Identity and Session Security

- 状态 / Status: `planned`
- 价值与风险 / Value and risk: 把本地认证骨架升级为真实 OIDC 安全边界。 / Turn the local authentication skeleton into a real OIDC security boundary.
- 启动条件 / Trigger: 确认托管 OIDC 厂商和正式域名。 / Managed OIDC provider and production domain are selected.
- 前置依赖 / Dependencies: issuer、audience、scope、退出和密钥轮换方案。 / Issuer, audience, scope, logout, and key-rotation design.
- 最小范围 / Minimum scope: 真实登录、token 验证、session 过期、撤销、CSRF/CORS 检查和审计。 / Real login, token validation, session expiry, revocation, CSRF/CORS checks, and audit.
- 验收证据 / Evidence: 过期/撤销/错误 issuer-audience 测试和真实浏览器登录记录。 / Expired, revoked, and wrong issuer-audience tests plus a real browser login record.

### 3. 备份、恢复与迁移演练 / Backup, Recovery, and Migration Drill

- 状态 / Status: `planned`
- 价值与风险 / Value and risk: 防止数据库或 migration 故障造成不可恢复的数据损失。 / Prevent unrecoverable loss from database or migration failure.
- 启动条件 / Trigger: 确认托管 PostgreSQL 和部署环境。 / Managed PostgreSQL and deployment environment are selected.
- 前置依赖 / Dependencies: migration ownership、备份目标和恢复权限。 / Migration ownership, backup target, and restore permissions.
- 最小范围 / Minimum scope: 自动备份、恢复到隔离环境、前向 migration、停止条件和回退限制。 / Automated backup, isolated restore, forward migration, stop conditions, and rollback limits.
- 验收证据 / Evidence: 一次计时恢复演练和数据一致性检查。 / One timed restore drill and consistency verification.

### 4. 基础指标与告警 / Baseline Metrics and Alerts

- 状态 / Status: `planned`
- 价值与风险 / Value and risk: 日志只能解释单次故障，指标用于发现积压、失败率和依赖退化。 / Logs explain individual failures; metrics reveal backlog, failure rate, and dependency degradation.
- 启动条件 / Trigger: API 和 worker 在测试环境稳定运行。 / API and worker run stably in a test environment.
- 前置依赖 / Dependencies: 指标命名、隐私字段和告警接收方。 / Metric naming, privacy fields, and alert destination.
- 最小范围 / Minimum scope: 请求延迟/错误率、Capture 队列深度、worker 失败、Pending Action 过期和数据库池指标。 / Request latency/error rate, Capture queue depth, worker failures, Pending Action expiry, and database-pool metrics.
- 验收证据 / Evidence: 故障注入触发告警并关联 request ID 日志。 / Fault injection triggers an alert correlated with request ID logs.

### 5. 真实部署与回滚 / Real Deployment and Rollback

- 状态 / Status: `planned`
- 价值与风险 / Value and risk: 把“可以构建”升级为“可以安全发布和恢复”。 / Upgrade from buildable to safely deployable and recoverable.
- 启动条件 / Trigger: 容器、数据库、OIDC、域名和 TLS 均已确认。 / Container, database, OIDC, domain, and TLS choices are confirmed.
- 前置依赖 / Dependencies: 不可变镜像、CI、Runbook、备份和审批。 / Immutable image, CI, Runbook, backup, and approval.
- 最小范围 / Minimum scope: 测试环境发布、健康验证、失败注入、上一镜像回滚和变更记录。 / Test deployment, health verification, fault injection, previous-image rollback, and change record.
- 验收证据 / Evidence: 发布与回滚时间线、日志、健康检查和用户侧验证。 / Deployment/rollback timeline, logs, healthchecks, and user-side verification.

## B. 触发后建设 / Trigger-Based Capabilities

### 6. 第一个 Source Subscription 连接器 / First Source Subscription Connector

- 状态 / Status: `trigger-based`
- 启动条件 / Trigger: 用户明确指定第一个来源、授权范围和期望刷新频率。 / The user specifies the first source, authorization scope, and refresh expectation.
- 最小范围 / Minimum scope: 只实现该来源；支持 pause、revoke、游标、幂等、限流和失败恢复，不建设通用抓取框架。 / Implement only that source with pause, revoke, cursor, idempotency, rate limits, and recovery; no generic ingestion framework.
- 验收证据 / Evidence: 重复抓取不重复创建 Capture，撤销后无新访问，既有证据按策略保留。 / Repeated ingestion creates no duplicate Capture, revocation stops access, and existing provenance follows policy.

### 7. 批量操作 / Bulk Operations

- 状态 / Status: `trigger-based`
- 启动条件 / Trigger: Inbox 或 Task 积压使逐项操作产生可测量负担。 / Inbox or Task backlog makes one-by-one actions measurably costly.
- 最小范围 / Minimum scope: 预览影响对象、创建一个有界 Pending Action、版本校验、部分失败策略和再次确认。 / Preview targets, create one bounded Pending Action, version-check, define partial-failure behavior, and reconfirm.
- 验收证据 / Evidence: 过期、目标变化、部分冲突和误操作恢复测试。 / Expiry, target-change, partial-conflict, and recovery tests.

### 8. Project paused / Project 暂停

- 状态 / Status: `trigger-based`
- 启动条件 / Trigger: 用户需要暂时停止新工作但保留项目在主要视图。 / The user needs to stop new work temporarily while keeping a Project visible.
- 最小范围 / Minimum scope: 明确 paused 对新增、移动、完成、显示和恢复的行为。 / Define paused behavior for creation, movement, completion, visibility, and resumption.
- 验收证据 / Evidence: 状态转换和所有 Task 边界测试。 / Transition and Task-boundary tests.

### 9. 独立 Label 实体 / Independent Label Entity

- 状态 / Status: `trigger-based`
- 启动条件 / Trigger: 出现跨 Task 重命名、颜色、统计、权限或大量重复值需求。 / Cross-Task rename, color, analytics, permissions, or substantial duplicate-value needs appear.
- 最小范围 / Minimum scope: Owner-scoped Label、唯一规范名、迁移和引用完整性。 / Owner-scoped Label, unique canonical name, migration, and referential integrity.
- 验收证据 / Evidence: 重命名、合并、并发和迁移测试。 / Rename, merge, concurrency, and migration tests.

### 10. 团队协作 / Team Collaboration

- 状态 / Status: `trigger-based`
- 启动条件 / Trigger: 第二个真实用户需要访问同一工作台。 / A second real user needs access to the same workbench.
- 最小范围 / Minimum scope: tenant、成员、角色、邀请、Project 权限和审计；先专项设计再改 schema。 / Tenant, members, roles, invitations, Project permissions, and audit; perform dedicated design before schema changes.
- 验收证据 / Evidence: 跨租户越权测试和权限矩阵。 / Cross-tenant denial tests and a permission matrix.

### 11. Redis 核心化 / Promote Redis into the Core

- 状态 / Status: `trigger-based`
- 启动条件 / Trigger: PostgreSQL worker 指标证明吞吐、延迟或协调无法满足目标。 / Metrics prove the PostgreSQL worker misses throughput, latency, or coordination goals.
- 最小范围 / Minimum scope: 明确 Redis 是队列还是协调层；PostgreSQL 仍保存业务状态；设计 outbox、重试、DLQ 和故障降级。 / Define Redis as queue or coordination; PostgreSQL keeps business state; design outbox, retry, DLQ, and outage behavior.
- 验收证据 / Evidence: 基准、双写故障、Redis 中断和恢复测试。 / Benchmarks, dual-write failure, Redis outage, and recovery tests.

### 12. 第三方任务平台同步 / Third-Party Task Platform Synchronization

- 状态 / Status: `trigger-based`
- 启动条件 / Trigger: 用户明确要求某个平台成为受控同步目标。 / The user explicitly requests a platform as a controlled synchronization target.
- 最小范围 / Minimum scope: 单个平台、方向明确、游标、幂等、冲突、撤销和删除语义；工作台仍是事实来源。 / One platform with explicit direction, cursor, idempotency, conflict, revocation, and deletion semantics; the workbench remains source of truth.
- 验收证据 / Evidence: 重放、乱序、冲突、断连和恢复测试。 / Replay, out-of-order, conflict, disconnect, and recovery tests.

## C. 当前拒绝 / Currently Rejected

### 13. 被动监控全部历史对话 / Passive Monitoring of All Conversation History

- 状态 / Status: `rejected`
- 原因 / Reason: 能力不可保证，且违反最小授权和最小数据保存边界。 / The capability cannot be guaranteed and violates least authorization and data minimization.

### 14. 为未知来源建设通用抓取框架 / Generic Ingestion Framework for Unknown Sources

- 状态 / Status: `rejected`
- 原因 / Reason: 没有具体来源契约时只能产生推测性抽象。 / Without a concrete source contract, it creates only speculative abstractions.

### 15. 微服务、第二后端或 Python Agent runtime / Microservices, Second Backend, or Python Agent Runtime

- 状态 / Status: `rejected`
- 原因 / Reason: 当前规模下增加重复契约、部署和故障边界，没有可测量收益。 / At current scale, these add duplicate contracts, deployment, and failure boundaries without measurable benefit.

## D. 路线图维护规则 / Roadmap Maintenance Rules

- 每项启动前重新验证触发条件，不因出现在本文件中而自动获批。 / Revalidate the trigger before starting; presence in this file is not automatic approval.
- 改动 API、schema、权限或外部系统前，新增或更新 ADR。 / Add or update an ADR before changing API, schema, authorization, or external integration boundaries.
- 完成项移动到变更记录，并附代码、测试、运行或演练证据。 / Move completed items to the change log with code, test, runtime, or drill evidence.
- 新想法必须说明真实问题和触发条件；没有触发条件则拒绝进入路线图。 / A new idea must state a real problem and trigger; without a trigger it does not enter the roadmap.
