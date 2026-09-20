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
