# Go Workbench

Workbench for turning captured inputs into reviewable tasks.

[简体中文](README.md) | [English](README.en.md)

[![Version](https://img.shields.io/badge/version-0.1.0--dev-orange)](openapi/openapi.yaml)
[![Status](https://img.shields.io/badge/status-work_in_progress-yellow)](#current-status)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1-6BA539?logo=openapiinitiative&logoColor=white)](openapi/openapi.yaml)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?logo=nextdotjs&logoColor=white)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=000000)](https://react.dev/)

## Overview

Go Workbench separates captured source material, machine-generated candidates, and tasks explicitly accepted by the user. Formal Tasks are the project's single source of truth, while source evidence and audit records remain separate.

## Current status

Version `0.1.0-dev` is not a release.

| Area | Status |
| --- | --- |
| Domain and architecture decisions | Documented |
| OpenAPI contract | Initial read-only Task contract available |
| Generated TypeScript API types | Available |
| Next.js workbench page | Minimal scaffold builds successfully |
| Go Task service | Interfaces, models, and a test scaffold exist; service behavior is incomplete |

## Repository layout

```text
.
├── docs/adr/                 Architecture decision records
├── internal/task/            Task domain and service package
├── openapi/openapi.yaml      API contract
├── README.md                 Chinese project documentation
└── web/                      Next.js frontend
```

## Development baseline

* Go 1.26
* Node.js 26
* npm 11

Install frontend dependencies:

```bash
npm --prefix web install
```

Generate the API types and validate the frontend:

```bash
npm --prefix web run generate:api
npm --prefix web run lint
npm --prefix web run typecheck
npm --prefix web run build
```

## Known limitations

- The Go HTTP API server and persistence layer are not implemented.
- `task.Service.List` currently returns `ErrNotImplemented`, so the Go test does not pass.

## Documentation

* [Architecture decisions](docs/adr/)
* [OpenAPI contract](openapi/openapi.yaml)

## Contributors

<p>
  <a href="https://github.com/MarnieWu" title="MarnieWu"><img src="https://images.weserv.nl/?url=github.com/MarnieWu.png%3Fsize%3D128&w=64&h=64&fit=cover&mask=circle" width="64" height="64" alt="MarnieWu" /></a>
  <a href="https://github.com/openai/codex" title="OpenAI Codex"><img src="https://images.weserv.nl/?url=github.com/openai.png%3Fsize%3D128&w=64&h=64&fit=cover&mask=circle" width="64" height="64" alt="OpenAI Codex" /></a>
</p>

## Security

Never commit keys, tokens, wallet files, database credentials, cookies, or real `.env` files. Use environment variables and commit only sanitized examples such as `.env.example` .

## License

No license has been selected. Do not assume permission to redistribute or reuse the code outside this repository.
