# Go Workbench

将捕获内容转化为可审核任务的工作台。

[简体中文](README.md) | [English](README.en.md)

[![Version](https://img.shields.io/badge/version-0.1.0--dev-orange)](openapi/openapi.yaml)
[![Status](https://img.shields.io/badge/status-work_in_progress-yellow)](#当前状态)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1-6BA539?logo=openapiinitiative&logoColor=white)](openapi/openapi.yaml)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?logo=nextdotjs&logoColor=white)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=000000)](https://react.dev/)

## 项目简介

Go Workbench 明确分离捕获的来源材料、机器生成的候选任务，以及用户明确接受的正式任务。正式 Task 是项目的单一事实来源，来源证据和审计记录独立保存。

## 当前状态

`0.1.0-dev` 不是正式发布版本。

| 范围 | 状态 |
| --- | --- |
| 领域与架构决定 | 已记录 |
| OpenAPI 契约 | 已有初始只读 Task 契约 |
| 生成的 TypeScript API 类型 | 已生成 |
| Next.js 工作台页面 | 最小脚手架构建通过 |
| Go Task service | 已有接口、模型和测试骨架，service 行为尚未完成 |

## 仓库结构

```text
.
├── docs/adr/                 架构决策记录
├── internal/task/            Task 领域与 service 包
├── openapi/openapi.yaml      API 契约
├── README.en.md              英文项目说明
└── web/                      Next.js 前端
```

## 开发基线

* Go 1.26
* Node.js 26
* npm 11

安装前端依赖：

```bash
npm --prefix web install
```

生成 API 类型并验证前端：

```bash
npm --prefix web run generate:api
npm --prefix web run lint
npm --prefix web run typecheck
npm --prefix web run build
```

## 已知限制

- Go HTTP API server 和持久化层尚未实现。
- `task.Service.List` 当前返回 `ErrNotImplemented`，因此 Go 测试尚未通过。

## 文档

* [架构决策记录](docs/adr/)
* [学习问答：Service、Repository、Context 与测试盲区](docs/learning-notes/01-service-repository-context-testing.md)
* [OpenAPI 契约](openapi/openapi.yaml)

## 贡献者

<p>
  <a href="https://github.com/MarnieWu" title="MarnieWu"><img src="https://images.weserv.nl/?url=github.com/MarnieWu.png%3Fsize%3D128&w=64&h=64&fit=cover&mask=circle" width="64" height="64" alt="MarnieWu" /></a>
  <a href="https://github.com/openai/codex" title="OpenAI Codex"><img src="https://images.weserv.nl/?url=github.com/openai.png%3Fsize%3D128&w=64&h=64&fit=cover&mask=circle" width="64" height="64" alt="OpenAI Codex" /></a>
</p>

## 安全

禁止提交密钥、令牌、钱包文件、数据库凭据、Cookie 或真实 `.env` 文件。使用环境变量，只提交经过脱敏的示例文件，例如 `.env.example` 。

## 许可证

项目尚未选择许可证。在许可证明确前，不要假设可以在仓库外重新分发或复用代码。
