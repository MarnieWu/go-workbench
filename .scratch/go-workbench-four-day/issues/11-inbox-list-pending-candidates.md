# 11 — 展示 pending Candidate 的 Inbox

**What to build:** 用户打开 Inbox 时，只看到当前 owner 的 `pending_review` Candidate；空状态说明还没有待审核建议，错误状态支持重试。

**Blocked by:** 09 — 幂等创建 Capture 的最小 API 切片。

**Status:** ready-for-agent

**Collaboration:** 我主写。你审查 owner scope、状态过滤和领域词是否准确；我实现查询、OpenAPI、页面状态和回归测试。

## 关键步骤

1. 写 repository RED 测试：owner scope、只返回 `pending_review`、已接受和已拒绝不出现在默认 Inbox。
2. 接通 `GET /v1/inbox` 的服务、HTTP 和 OpenAPI；先允许测试直接插入 Candidate，不要求 worker 已实现。
3. Web 增加 Inbox 视图，使用生成的客户端读取数据。
4. 补 skeleton、loading、empty、error、success 五种状态。

## 改动范围

- Domain：Candidate 列表查询模型。
- Data：Candidate 只读查询，不修改 Capture。
- API：Inbox 读取端点和响应契约。
- UI：Inbox 页面和状态组件。
- Tests：repository、router、generated client drift、Web typecheck。

- [ ] Owner A 看不到 Owner B 的 Candidate。
- [ ] `accepted` 和 `rejected` Candidate 不出现在默认 Inbox。
- [ ] 空结果返回 `items: []`，不是 `null`。
- [ ] Inbox 使用生成的 OpenAPI 客户端，不手写第二套响应类型。
