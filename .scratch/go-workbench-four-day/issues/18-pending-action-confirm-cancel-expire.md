# 18 — 确认、取消和过期 Pending Action

**What to build:** 用户确认 Pending Action 时，系统重新检查目标 version 和过期时间；有效动作执行，取消和过期动作不修改 Task。

**Blocked by:** 17 — 完成或归档只创建 Pending Action。

**Status:** ready-for-agent

**Collaboration:** 你主写。你实现 Pending Action 状态机、clock 注入和事务确认；我提供失败矩阵、UI mutation、测试夹具和审查。

## 关键步骤

1. 写有效确认、目标已变化、过期、取消、重复确认和执行失败的 RED 测试。
2. 用户实现事务状态机；时间通过可替换 clock 注入，测试不能真实等待。
3. 接通 confirm、cancel API 和 Web mutation。
4. 成功确认后刷新 Task 与 Pending Action 视图。

## 改动范围

- Domain：Pending Action 状态机。
- Data：目标 recheck、Task mutation、Pending Action 状态和 Audit Event 同事务。
- API：confirm/cancel endpoints。
- UI：确认、取消、失败恢复。
- Tests：事务、clock、router、Web state。

- [ ] stale、expired、cancelled action 不修改 Task。
- [ ] 有效确认原子修改 Task、标记 executed、追加 Audit Event。
- [ ] 重复确认不会重复执行。
- [ ] 失败状态与取消、过期、已执行可区分。
