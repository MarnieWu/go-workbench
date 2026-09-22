# 11 — 用 version 防止 Task 被旧数据覆盖

**结果：** 更新 Task 时必须提交当前 version；旧 version 返回 `409 VERSION_CONFLICT`，数据库内容不变。

**Blocked by:** 08。

**Status:** ready-for-agent

## 步骤与验证

1. 写三个测试：正确 version 成功、旧 version 冲突、两个并发相同 version 最多一个成功。
   - 验证：`go test ./... -run 'Test.*Task.*Version' -count=1`。
   - RED 通过：失败原因是 version 规则未实现。
2. 你写带 `owner_id + id + version` 条件的 UPDATE，成功后 version 加一。
   - 核对：不能先查再无条件更新；要检查 affected rows。
3. 重跑目标测试并查询数据库。
   - 通过：旧请求未改任何字段；成功响应含新 version。
4. 写 router 测试检查 `409 VERSION_CONFLICT`。
   - 通过：状态码和稳定 code 正确，用户输入可在 UI 保留。
5. 运行完整 tests 和 race。
   - 通过：全部 `ok`，无 DATA RACE。

## 最终通过

- [ ] 所有 Task 写入都带 owner、ID、version 条件。
- [ ] 旧 version 不覆盖新数据。
- [ ] 并发写最多一个成功。
- [ ] reopen 与 restore 仍是不同动作。
