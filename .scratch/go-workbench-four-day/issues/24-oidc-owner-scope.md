# 24 — OIDC Owner 身份和权限边界

**What to build:** API 验证托管 OIDC token 的签名、issuer、audience、expiration 和 scope，只用可信 subject 映射 Owner；测试 issuer/JWKS 不能使用真实 token。

**Blocked by:** 20 — 第二阶段一致性总验收。

**Status:** ready-for-agent

**Collaboration:** 你主写。你实现 token 校验、Owner 映射和 401/403 边界；我搭测试 issuer/JWKS、负向矩阵和日志脱敏检查。

## 关键步骤

1. Agent 建测试 issuer/JWKS；用户写缺失、签名错误、过期、错误 issuer/audience、缺 scope 和合法 token 测试。
2. 用户实现 middleware 和 claim 校验；不能读取客户端 owner header/body。
3. 替换本地 owner 注入的默认运行路径，保留显式开发模式。
4. 检查日志不含 token、Cookie、完整 claims 或 secret。

## 改动范围

- Auth：OIDC middleware、Owner resolver、scope 校验。
- API：401/403 边界。
- Data：owner mapping。
- Tests：auth negative matrix、owner scope、log redaction。

- [ ] 五类非法 token 均被拒绝。
- [ ] 401 与 403 不混用。
- [ ] Owner 只来自验证后的 subject。
- [ ] 没有真实 OIDC 环境时，live OIDC 验证标记为 blocked。
