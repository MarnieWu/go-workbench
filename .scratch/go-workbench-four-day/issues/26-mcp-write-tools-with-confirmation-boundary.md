# 26 — MCP 写工具和确认边界

**What to build:** MCP 支持明确创建 Task、版本更新 Task，以及只创建 Pending Action 的 complete/archive；不提供删除、批量或自动建 Project。

**Blocked by:** 25、18。

**Status:** ready-for-agent

**Collaboration:** 你主写。你实现写工具的明确指令判断、version 检查和确认边界；我主写 schema、调用测试和拒绝矩阵。

## 关键步骤

1. 写工具权限矩阵：明确指令 create/update 可执行，高风险 complete/archive 只能 propose。
2. 用户实现 handler 权限判断和版本检查；Agent 补 schema 和调用测试。
3. complete/archive 工具调用后，Task 当场不变化。
4. 检查日志和工具输出不保存完整会话正文。

## 改动范围

- MCP：create/update/complete/archive tools。
- Domain：复用 Task 和 Pending Action 服务。
- Auth：明确指令和 owner scope。
- Tests：允许/拒绝矩阵、pending boundary。

- [ ] 没有 delete 或 bulk MCP 工具。
- [ ] create Task 写 Audit Event。
- [ ] complete/archive 只创建 Pending Action。
- [ ] 工具不能绕过 version 和 owner scope。
