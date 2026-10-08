# 30 — 发布候选、Runbook 和最终验收

**What to build:** 得到可复现的非 root 镜像、核心 Compose、CI、扫描结果、Runbook 和逐项验收记录；外部发布、生产 migration、Swarm 写入不在本票自动执行。

**Blocked by:** 29 — 浏览器到 Worker 的端到端与 request ID 排障。

**Status:** ready-for-agent

**Collaboration:** 我主写。你确认发布阈值、停止条件和外部环境授权；我实现 Docker/Compose/CI/Runbook/验收清单整理，不执行外部发布或生产变更。

## 关键步骤

1. 构建多阶段镜像，确认最终容器非 root。
2. 验证 Compose config、完整启动、healthcheck、数据库不可用时 readiness。
3. 建立 CI，覆盖格式、vet、tests/race、契约漂移、Web lint/typecheck/test/build 和扫描。
4. 运行 govulncheck 和镜像扫描；阈值未确认时标 blocked。
5. 写 RUNBOOK：前置条件、脱敏配置、migration 前检查、启动、观察、停止、恢复、备份、回退。
6. 按验收清单逐项填状态和证据。

## 改动范围

- Delivery：Dockerfile、Compose、CI。
- Security：扫描和处置记录。
- Docs：RUNBOOK、验收清单、最终限制。
- Tests：最终 gate 命令。

- [ ] 核心镜像非 root，Compose healthy。
- [ ] Go、Web、E2E、race、契约漂移检查通过。
- [ ] 扫描结果有阈值或明确 blocked 说明。
- [ ] RUNBOOK 可由无上下文新会话执行。
- [ ] 未将本地验证描述为生产验证。
