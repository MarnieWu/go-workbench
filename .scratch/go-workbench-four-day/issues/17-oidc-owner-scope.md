# 17 — 验证登录令牌并得到 Owner

**结果：** API 验证签名、issuer、audience、expiration 和 scope，只使用可信 subject 映射 Owner。

**Blocked by:** 13。

**Status:** ready-for-agent

## 步骤与验证

1. 让 Agent 建测试 issuer/JWKS，不使用真实 token。你写缺失、签名错误、过期、错误 issuer/audience、缺 scope 和合法 token 测试。
   - 验证：`go test ./... -run 'Test.*Auth' -count=1` 先 RED。
2. 你实现 middleware 和 claim 校验。
   - 核对：Owner 只来自验证后的 subject，不读取客户端 owner header/body。
3. 重跑测试。
   - 通过：未认证返回 401，缺权限返回 403，合法 token 进入 handler。
4. 运行两个 Owner 的越权测试。
   - 通过：Owner A 无法读取或修改 Owner B 数据。
5. 检查日志捕获测试。
   - 通过：日志没有原 token、Cookie、完整 claims 或 secret。
6. 运行完整 tests/race。
   - 通过：全部 `ok`。

## 最终通过

- [ ] 五项 token 条件都验证。
- [ ] 401 与 403 不混用。
- [ ] SQL 仍显式 owner scope。
- [ ] 没有真实身份环境时，live OIDC 保持 blocked。
