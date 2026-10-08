# 25 — MCP 的 capture 和只读工具

**What to build:** Go MCP server 暴露 `capture_todo`、`list_projects`、`list_tasks`，工具有 owner scope、分页/长度限制和安全错误，不执行外部业务动作。

**Blocked by:** 21、24。

**Status:** ready-for-agent

**Collaboration:** 共同完成。你主写 MCP 权限判断和 owner scope；我主写 schema、server 接线、输出大小限制和只读工具测试。

## 关键步骤

1. 列出每个工具的输入、输出、最大长度、权限和禁止行为。
2. Agent 接 MCP server 和 schema 测试；用户实现 handler 与权限判断。
3. `capture_todo` 只创建 Capture，不创建 Candidate 或 Task。
4. 只读工具必须分页，输出大小受控。

## 改动范围

- MCP：server 入口、tool schema、handler。
- Domain：复用已有 capture/list 服务。
- Auth：工具 owner scope。
- Tests：schema、允许/拒绝矩阵、输出大小。

- [ ] MCP 不复制业务规则。
- [ ] `capture_todo` 不直接创建 Task。
- [ ] list 工具有分页和 owner scope。
- [ ] 错误不泄露内部信息或完整会话。
