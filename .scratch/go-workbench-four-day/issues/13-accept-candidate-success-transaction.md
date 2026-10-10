# 13 — 接受 Candidate 的成功事务

**What to build:** 用户编辑 Candidate 标题、描述、label 和 Project 建议后接受；系统在一个事务中创建 Task、关联 Source Evidence、标记 Candidate accepted，并追加 Audit Event。

**Blocked by:** 11 — 展示 pending Candidate 的 Inbox。

**Status:** waiting-for-user

**Agent progress (2026-10-10):** accept HTTP/OpenAPI 接线、字段编辑 UI 和 happy-path router 测试已完成。`pgx.Tx`、Task/Evidence/Candidate/Audit 原子写入仍由用户实现，运行入口因此尚未注入 accept service。

**Collaboration:** 你主写。你实现 `pgx.Tx` 成功事务、字段映射和核心测试；我提供 RED 测试框架、HTTP/OpenAPI/UI 接线和代码审查。

## 关键步骤

1. 写成功路径 RED 测试，明确事务结束后四类记录的期望状态。
2. 用户实现 `pgx.Tx` 成功事务；Agent 提供 HTTP/OpenAPI/UI 接线和测试夹具。
3. Web 接受前允许编辑 Candidate 建议字段。
4. 接受完成后跳转或刷新 Task 列表，新增 Task 可见。

## 改动范围

- Domain：Candidate accept 输入、Task 创建字段、label 复制规则。
- Data：Task、Task Evidence Link、Candidate、Audit Event 同事务写入。
- API：accept endpoint 和响应。
- UI：Candidate edit + accept mutation。
- Tests：事务成功、router、Web happy path。

- [ ] 一个成功接受动作创建恰好一个 Task。
- [ ] Candidate 状态变为 `accepted`，并保存 accepted task ID。
- [ ] Task 与 Source Evidence 建立唯一关联。
- [ ] Audit Event 与业务写入一起提交。
