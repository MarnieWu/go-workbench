# 21 — 跑通浏览器到 Worker 的完整流程

**结果：** 在浏览器完成 Capture、Inbox、接受 Candidate、查看 Task、确认 Pending Action，并用 request ID 找到对应 API/Worker 日志。

**Blocked by:** 20。

**Status:** ready-for-agent

## 你负责什么

你负责关键 Go 失败测试、故障注入和排障判断。Agent 可以写 Playwright、UI 状态和样式胶水。

## 步骤与验证

1. 让 Agent 创建 Playwright 用例空壳，按业务顺序填写：提交 Capture → 等待 Candidate → 接受 → 查看 Task → propose → confirm。
   - 验证：先运行 E2E，因流程尚未接完而 RED；测试环境错误不算业务 RED。
2. 逐段接通流程，每完成一段只重跑对应 E2E。
   - 通过：页面结果与数据库状态一致，不能只断言页面文字出现。
3. 增加 stale version、过期 Action、API 失败和 Worker 失败场景。
   - 通过：页面保留输入、显示安全重试，不出现内部错误。
4. 检查 Inbox、Task、Project、Pending Action 的 loading/empty/error/success。
   - 通过：每种状态都可稳定复现，不靠临时修改生产代码。
5. 在 360、768、1024px 检查布局，并只用键盘完成核心流程。
   - 通过：无水平滚动、焦点可见、按钮有可访问名称。
6. 从浏览器 Network 复制 request ID，在 API 和 Worker JSON 日志查询。
   - 通过：能定位同一业务流程，日志无完整正文和敏感字段。
7. 运行：
   ```bash
   go test ./... -count=1
   go test -race ./... -count=1
   npm --prefix web run lint
   npm --prefix web run typecheck
   npm --prefix web run test
   npm --prefix web run build
   npm --prefix web run test:e2e
   ```
   - 通过：所有实际存在的命令退出码为 0；缺少脚本必须先补或标明 blocked，不能跳过后声称通过。

## 最终通过

- [ ] 主流程 E2E 通过并核对数据库结果。
- [ ] 关键失败流程 E2E 通过。
- [ ] 三个 viewport 和键盘流程无阻断。
- [ ] request ID 可定位 API/Worker 日志。
- [ ] 保存一次 Playwright trace 和一次故障恢复记录。
