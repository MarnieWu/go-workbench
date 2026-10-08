# 28 — Worker、Auth、MCP 运行门禁

**What to build:** API、Worker 和 MCP 可独立启动和停止；Worker 并发、重试、shutdown、OIDC、MCP 和 Redis 实验边界都有可复现证据。

**Blocked by:** 23、24、26、27。

**Status:** ready-for-agent

**Collaboration:** 共同完成。我组织运行门禁、记录证据和定位失败；你确认状态机、权限边界和是否允许进入端到端阶段。

## 关键步骤

1. 运行 Go test、race、vet、govulncheck 或记录工具缺失。
2. 启动 API、Worker、MCP 三个入口，逐个验证健康、停止和日志。
3. 验证 Redis profile 关闭后核心仍通过。
4. 更新验收清单证据。

## 改动范围

- Runtime：启动脚本或 Make target。
- Tests：聚合 gate。
- Docs：验证记录和 blocked 项。
- Fixes：只修复 gate 发现的最小问题。

- [ ] 并发 Worker 不重复处理同一 Capture。
- [ ] Auth 负向用例全部拒绝。
- [ ] MCP 高风险工具不直接修改 Task。
- [ ] 日志不含 token、Cookie、完整正文或 secret。
