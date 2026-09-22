# 08 — 从 PostgreSQL 查询当前 Owner 的 Task

**结果：** 原来的 Task API 改为从 PostgreSQL 读取，Owner A 永远看不到 Owner B 的数据。

**Blocked by:** 07。

**Status:** ready-for-agent

## 步骤与验证

1. 在隔离数据库插入两个 Owner 及各自 Task fixture。
2. 先写 Repository 集成测试：Owner A 查询只返回 A；空 Owner 返回空列表；status filter 正确。
   - 验证：运行 `go test ./... -run 'Test.*Repository.*List' -count=1`。
   - RED 通过：测试因查询未实现失败，不因数据库未启动失败。
3. 你写参数化 SQL、`rows.Scan` 和错误处理。SQL 必须显式带 `owner_id` 条件。
   - 核对：没有字符串拼接 SQL；`Rows` 被关闭；请求 context 被传给 pgx。
4. 重跑 Repository 测试。
   - 通过：全部 `ok`，测试明确断言没有 Owner B 数据。
5. 运行真实 router 测试和 `go test ./... -count=1`。
   - 通过：HTTP JSON 不变，完整测试无 FAIL。
6. 在浏览器查看 Task 页面。
   - 通过：页面显示数据库数据，Network 响应与 Repository 测试一致。

## 最终通过

- [ ] SQL 显式 owner scope。
- [ ] 两个 Owner 隔离测试通过。
- [ ] 空列表、status filter 和 router 测试通过。
- [ ] 没有把 SQL 写进 handler。
