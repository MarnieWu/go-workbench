# 09 — 幂等创建 Capture 的最小 API 切片

**What to build:** 用户从 Web 或 HTTP 提交一段可见 TODO 内容后，系统创建一条 `queued` Capture 和最小 Source Evidence；重复提交同一 owner、同一幂等键、同一输入时返回原 Capture，不创建 Candidate 或 Task。

**Blocked by:** 08 — 让 Task 列表从 PostgreSQL 读取。

**Status:** ready-for-agent

**Collaboration:** 共同完成。你主写输入 hash、幂等规则、repository 冲突处理和核心测试；我主写 OpenAPI、HTTP 接线、Web 表单状态和测试脚手架。

## 关键步骤

1. 先写服务和 repository 的 RED 测试：首次提交成功、重复提交返回同一 Capture、不同输入返回冲突、不同 owner 可复用同一幂等键。
2. 用户实现输入 hash、幂等键约束处理和业务错误；Agent 只补脚手架、OpenAPI 响应形状和测试胶水。
3. 接通 `POST /v1/captures`，响应包含 Capture ID、状态和 request ID；错误响应使用稳定 code。
4. Web 增加一个最小“加入工作台”入口，提交期间禁用重复点击，失败时保留输入。

## 改动范围

- Domain：新增 Capture create 输入、输出和业务错误。
- Data：复用现有 Capture 与 Source Evidence 表；不新增自动清理。
- API：新增 capture 创建端点和 OpenAPI 契约。
- UI：新增最小提交入口和 loading/error 状态。
- Tests：服务、PostgreSQL 集成、真实 Gin router、前端类型检查。

- [ ] 同一 owner、同一幂等键、同一输入只创建一条 Capture。
- [ ] 同一幂等键、不同输入返回 `409 IDEMPOTENCY_CONFLICT`，响应不泄露另一份输入正文。
- [ ] Capture 初始状态为 `queued`，Candidate 和 Task 数量不变。
- [ ] Web mutation 期间不能重复提交，失败后输入仍在页面上。
- [ ] 验证命令至少覆盖 Go 目标测试、OpenAPI 生成漂移和 Web typecheck。
