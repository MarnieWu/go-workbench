# 18 — 用 Go 暴露受控 MCP 工具

**结果：** MCP 提供 capture、有限查询、明确 create Task，以及只创建 Pending Action 的 complete/archive。

**Blocked by:** 12、14、17。

**Status:** ready-for-agent

## 步骤与验证

1. 先列出每个工具的输入、输出、最大长度、所需权限和禁止行为。
   - 通过：没有删除、批量、自动建 Project 或任意外部动作。
2. 让 Agent 创建 MCP server 接线和 schema 测试，你写 handler 与权限判断。
   - 验证：schema/handler 测试先 RED。
3. 实现 `capture_todo`。
   - 通过：只创建 Capture，不创建 Candidate/Task。
4. 实现 list/create 工具。
   - 通过：list 有分页/数量限制和 owner scope；create Task 需要明确指令并写 Audit Event。
5. 实现 complete/archive。
   - 通过：只创建 Pending Action，Task 当场不变化。
6. 运行工具允许/拒绝矩阵和完整 tests/race。
   - 通过：全部 `ok`，错误不泄露内部信息。

## 最终通过

- [ ] MCP 复用已有 Service，不复制业务规则。
- [ ] 工具不能绕过 Owner 和确认边界。
- [ ] 输出大小受控，日志不保存完整会话。
- [ ] 未接真实 ChatGPT 时明确标记 blocked。
