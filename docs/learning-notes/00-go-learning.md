# Memos

Go 相关

```go
// context 是调用链的生命周期管理工具
// controlling timeouts
// cancelling go routines
// and passing metadata across your Go application
import ("context")
```

项目初始化

```bash
// 根据 OpenAPI 生成 TypeScript 类型
npm --prefix web run generate:api
// 生成文件：web/lib/api/generated/schema.d.ts

// 验证前端脚手架
npm --prefix web run lint
npm --prefix web run typecheck
npm --prefix web run build

// 安装依赖
npm --prefix web install
// 它会创建或更新：
// web/node_modules/
// web/package-lock.json

// 创建 Go 模块
go mod init go-workbench
// 它会生成：go.mod

// 创建 Next.js 脚手架
npx create-next-app@latest web \
  --typescript \
  --eslint \
  --tailwind \
  --app \
  --use-npm
```

常见目录结构

```go
my-project/
├── go.mod    // 声明模块、Go 版本和直接依赖，它是 Go module 的入口
├── go.sum    // 记录依赖内容的校验值，由 Go 工具自动维护。通常需要提交，不要手工编辑。
├── cmd/    // 存放可执行程序入口
│   ├── api/
│   │   └── main.go
│   └── worker/
│       └── main.go
├── internal/   // 存放项目内部代码，其他 Go module 不能导入这里的 package。这是 Go 工具实际执行的访问限制
│   ├── task/
│   ├── httpapi/
│   └── repository/
├── web/    // Next.js 前端
├── migrations/   // 存放数据库结构变更
├── openapi/    // 存放 HTTP API 契约
├── testdata/   // 存放测试输入文件，例如 JSON、图片或 SQL 样本。Go 工具会忽略名为 testdata 的目录，不会把它当普通 package 构建。
├── README.md
└── Makefile
```

cmd 下存放可执行程序入口，即 cmd 下任何一个 package 都是可以独立启动的程序：

```go
go run ./cmd/api  // api 程序
go run ./cmd/worker   // worker 程序
```

程序启动后，就可以生成对应的程序进程。

internal 下放不能独立启动的共享代码，它们更多是被其他程序调用：

```go
api 进程
   ↓
task.Create()

worker 进程
   ↓
task.Create()
```

所以：

```go
cmd/api       → 一个程序入口
cmd/worker    → 另一个程序入口
internal/task → 两个程序复用的普通代码
```

Gin 框架
