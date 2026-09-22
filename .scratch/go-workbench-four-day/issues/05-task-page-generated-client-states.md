# 05 — 用生成客户端显示 Task 页面

**结果：** 页面只用 OpenAPI 生成客户端请求 Task，并显示 loading、empty、success、error/retry。

**Blocked by:** 03、04。

**Status:** ready-for-agent

## 你负责什么

你负责核对 Go 响应和生成类型是否一致；Agent 可以写 UI 和样式胶水。不要手写第二套 Task API 类型。

## 步骤与验证

1. 运行 `npm --prefix web run generate:api`。
   - 核对：命令退出码为 0。
   - 通过：生成文件能找到 `listTasks` 对应类型。
2. 运行 `git diff -- web/lib/api/generated`。
   - 核对：契约未变时应无意外 diff；有 diff 时逐项解释来源。
   - 通过：没有无法解释的漂移。
3. 接入 success 和 empty 状态，再模拟 loading 和 API error。
   - 核对：空数据展示说明和下一步；错误展示重试及 request ID。
   - 通过：四种状态均能稳定复现。
4. 在 360px、768px、1024px viewport 检查页面。
   - 核对：无水平滚动，按钮可见，键盘可到达重试按钮。
   - 通过：三个宽度无阻断。
5. 运行：
   ```bash
   npm --prefix web run lint
   npm --prefix web run typecheck
   npm --prefix web run build
   ```
   - 核对：三条命令退出码都为 0。
   - 通过：没有 error；warning 必须记录并判断。

## 最终通过

- [ ] 页面没有手写重复 API 响应类型。
- [ ] loading、empty、success、error/retry 都可复现。
- [ ] 三个 viewport 无阻断。
- [ ] lint、typecheck、build 全部通过。

完成后解释：TypeScript 编译通过为什么不能替代 Go router 测试。
