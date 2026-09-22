# 15 — 防止两个 Worker 重复处理同一 Capture

**结果：** 多 Worker 可并行，但同一 Capture 同时只能被一个 Worker 领取；失败重试有限，人工重放不创建新 Capture。

**Blocked by:** 14。

**Status:** ready-for-agent

## 步骤与验证

1. 写并发测试：两个 Worker 同时竞争同一 queued Capture。
   - 验证：目标测试先 RED，不能用 sleep 假装同步。
2. 你实现数据库原子领取和受控并发。
   - 通过：测试数据库中只有一次处理结果。
3. 写 transient、permanent、达到重试上限三个测试。
   - 通过：临时错误有限重试；永久错误不空转；达到上限进入 failed。
4. 写人工重放测试。
   - 通过：Capture ID 和幂等身份不变，只增加 attempt/重新进入允许状态。
5. 运行目标测试 10 次，再运行 race：
   ```bash
   go test ./... -run 'Test.*Worker.*Concurrent' -count=10
   go test -race ./... -count=1
   ```
   - 通过：十次均 `ok`，无 DATA RACE。

## 最终通过

- [ ] 同一 Capture 不被重复处理。
- [ ] retry/backoff 有上限。
- [ ] failed 不会无限循环。
- [ ] replay 不创建重复 Capture。
