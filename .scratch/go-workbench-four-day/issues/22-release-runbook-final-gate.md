# 22 — 构建发布候选并完成总验收

**结果：** 得到可复现的非 root 镜像、核心 Compose、CI、扫描结果、Runbook 和逐项验收记录。外部发布不在本任务中自动执行。

**Blocked by:** 21。

**Status:** ready-for-agent

## 步骤与验证

1. 让 Agent 起草多阶段 Dockerfile，你核对运行用户、二进制和启动命令。
   - 验证：`docker build -t workbench:local .`。
   - 通过：构建退出码 0，最终容器用户不是 root。
2. 验证 Compose：
   ```bash
   docker compose -f deploy/compose.yaml config
   docker compose -f deploy/compose.yaml up --build -d
   docker compose -f deploy/compose.yaml ps
   ```
   - 通过：核心服务 healthy；数据库不可用时 readiness 为 503，但 liveness 仍反映进程存活。
3. 发送终止信号并观察 API/Worker。
   - 通过：在配置超时内退出，无新任务领取，重启后 queued 工作继续。
4. 建立 CI，覆盖 format check、vet、Go tests/race、客户端漂移、Web lint/typecheck/test/build。
   - 验证：本地逐条执行与 CI 相同命令。
   - 通过：全部退出码为 0，CI 不依赖开发机隐藏状态。
5. 运行 `govulncheck ./...` 和 `trivy image workbench:local`。
   - 通过：达到事先确认的阈值；否则每个发现有处置。阈值未确认时状态是 blocked。
6. 验证可选实验。
   - Swarm 只运行静态 `docker stack config`；Nginx/Loki/Grafana 只在已有环境验证。
   - 通过：关闭实验后核心 build/test/Compose 仍成功；没有真实环境则标 blocked。
7. 写 RUNBOOK：前置条件、脱敏配置、migration 前检查、启动、观察、停止、恢复、备份和回滚。
   - 验证：让一个没有上下文的新会话只看 RUNBOOK，能复述步骤和停止条件。
8. 按 `GO_WORKBENCH_ACCEPTANCE_CHECKLIST.md` 逐项填写状态。
   - 通过：每个 passed 项附命令/测试/日志/截图；failed 项修复后重验；blocked 项写明缺少什么。

## 最终通过

- [ ] 核心镜像非 root，Compose healthy。
- [ ] Go、Web、E2E、race、契约漂移检查通过。
- [ ] 扫描结果有明确阈值和处置。
- [ ] 完成一次受控故障注入和恢复。
- [ ] Runbook 可由新会话执行。
- [ ] 未将“本地验证”描述为“已在生产验证”。

未经再次确认，不执行外部发布、生产 migration、Swarm stack 写入或不可逆操作。
