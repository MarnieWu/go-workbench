# 10 — 收紧 Capture 输入校验和安全错误

**What to build:** Capture 创建入口拒绝空内容、超长内容、未知 source、缺失幂等键和非法 JSON，并返回可预测的 4xx；日志和响应不包含完整正文、Authorization、Cookie 或真实配置。

**Blocked by:** 09 — 幂等创建 Capture 的最小 API 切片。

**Status:** waiting-for-user

**Agent progress (2026-10-10):** 已完成 HTTP 负向测试、日志脱敏测试、可注入测试 logger、Web 长度提示和失败保留输入。RED 测试当前精确阻塞在用户负责的未知字段、空/超长正文、缺幂等键和 source allowlist 校验。

**Collaboration:** 共同完成。你主写校验边界和错误映射判断；我主写负向 HTTP 测试、日志脱敏测试、Web 错误状态和重复样板。

## 关键步骤

1. 写 HTTP RED 测试覆盖非法 JSON、未知字段、空内容、超长内容、缺幂等键、非法 source。
2. 用户实现校验规则和错误映射；Agent 可以整理错误响应 helper，不能把内部 error 文本直接返回给客户端。
3. 加入日志捕获测试，确认 request ID 存在且敏感字段未写入。
4. 补 Web 的输入长度提示和失败恢复，不加 analytics 或外部网络调用。

## 改动范围

- API：请求绑定、字段限制、稳定错误 code。
- Domain：Capture 输入大小、source 类型和幂等键校验。
- Logs：allowlist 字段和 request ID。
- UI：表单错误、保留输入、安全重试。
- Tests：HTTP negative cases、日志脱敏、Web 状态。

- [ ] 所有非法输入在调用 repository 前返回 4xx。
- [ ] panic 和 repository error 仍返回安全 500。
- [ ] 日志不记录完整输入、token、Cookie 或连接字符串。
- [ ] Web 错误状态可复现，且不清空用户输入。
