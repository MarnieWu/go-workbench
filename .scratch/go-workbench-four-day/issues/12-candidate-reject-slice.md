# 12 — 拒绝 Candidate 并保留最小证据

**What to build:** 用户在 Inbox 拒绝一个 Candidate 后，它离开默认 Inbox，但 Capture、Source Evidence 和 Audit Event 保留可追溯记录。

**Blocked by:** 11 — 展示 pending Candidate 的 Inbox。

**Status:** waiting-for-user

**Agent progress (2026-10-10):** reject HTTP/OpenAPI 接线、稳定 404/409/500 映射、Web mutation 禁用和失败保留已完成。Candidate 状态更新、Audit Event 和事务仍由用户实现，运行入口因此尚未注入 reject service。

**Collaboration:** 共同完成。你主写 Candidate 状态规则和 Audit Event 写入判断；我主写 HTTP、UI mutation、错误状态和辅助测试。

## 关键步骤

1. 写服务和 repository RED 测试：拒绝 pending 成功、重复拒绝保持幂等或返回稳定冲突、跨 owner 拒绝失败。
2. 用户实现状态更新和 Audit Event 追加；Agent 只接线 HTTP、OpenAPI 和 UI。
3. Web 在拒绝中禁用按钮；失败后 Candidate 仍留在原位。
4. 重查 Inbox，确认被拒绝项不再出现。

## 改动范围

- Domain：Candidate reject 规则和错误。
- Data：Candidate 状态更新、Audit Event append-only。
- API：reject 端点和稳定错误 code。
- UI：Inbox reject mutation 和错误恢复。
- Tests：事务结果、owner scope、router、Web 状态。

- [ ] 拒绝后 Candidate 状态为 `rejected`，默认 Inbox 不再展示。
- [ ] Capture 和 Source Evidence 没有被改写或删除。
- [ ] Audit Event 记录 action、entity 和 request ID。
- [ ] 重复或非法状态不会创建不一致记录。
