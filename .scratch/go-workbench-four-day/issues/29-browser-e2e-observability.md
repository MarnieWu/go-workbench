# 29 — 浏览器到 Worker 的端到端与 request ID 排障

**What to build:** 用户在浏览器提交 Capture、查看 Inbox、接受 Candidate、查看 Task、propose 并 confirm Pending Action；同一流程能通过 request ID 定位 API 和 Worker 日志。

**Blocked by:** 28 — Worker、Auth、MCP 运行门禁。

**Status:** ready-for-agent

**Collaboration:** 我主写。你定义真实业务路径和判定失败是否可接受；我实现 Playwright、viewport、键盘流程、request ID 排障记录和 UI 修复。

## 关键步骤

1. Agent 建 Playwright 用例空壳；用户补关键失败判断和故障注入预期。
2. 逐段接通主流程，每段核对数据库状态，不能只看页面文字。
3. 增加 stale version、过期 action、API 失败和 Worker 失败场景。
4. 在桌面、平板、手机 viewport 和键盘流程下验证。
5. 从浏览器 Network 拿 request ID，查询 API 与 Worker JSON 日志。

## 改动范围

- UI：主流程、loading/empty/error/success、可访问性。
- E2E：Playwright 主路径和失败路径。
- Logs：request ID 查询步骤。
- Docs：故障恢复记录和 trace 路径。

- [ ] 主流程 E2E 通过并核对数据库。
- [ ] 关键失败流程保留输入并安全重试。
- [ ] 360、768、1024px 无水平滚动或遮挡。
- [ ] request ID 可定位同一业务流程的 API 和 Worker 日志。
