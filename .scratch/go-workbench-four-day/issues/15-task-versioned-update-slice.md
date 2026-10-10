# 15 — 用 version 更新 Task

**What to build:** 用户修改 Task 标题、描述、状态、优先级、label 或 Project 时必须提交当前 version；旧 version 返回冲突，不覆盖新数据。

**Blocked by:** 08 — 让 Task 列表从 PostgreSQL 读取。

**Status:** waiting-for-user

**Agent progress (2026-10-10):** PATCH HTTP/OpenAPI 接线、稳定 `VERSION_CONFLICT`、generated-client Web 编辑和冲突输入保留已完成。owner/ID/version 条件更新由用户实现，运行入口因此尚未注入 updater。

**Collaboration:** 你主写。你实现带 owner、ID、version 条件的更新和冲突测试；我主写 OpenAPI、router 胶水、Web 冲突状态和审查。

## 关键步骤

1. 写更新成功、旧 version 冲突、跨 owner 失败和并发同 version 的 RED 测试。
2. 用户实现带 owner、ID、version 条件的更新；不能先查再无条件写。
3. 接通 `PATCH /v1/tasks/{id}` 和 OpenAPI。
4. Web 提供最小编辑入口，冲突时保留本地输入。

## 改动范围

- Domain：Task update 输入和字段级校验。
- Data：乐观并发更新和 version 自增。
- API：PATCH endpoint、409 code、响应新 version。
- UI：Task 编辑、loading、conflict 状态。
- Tests：repository、router、Web typecheck、race。

- [ ] 成功更新后 version 增加。
- [ ] 旧 version 不改变数据库内容。
- [ ] 并发同 version 最多一个成功。
- [ ] Web 在冲突后不丢失用户输入。
