import { TaskList } from "./task-list";

export default function Home() {
  return (
    <main className="mx-auto min-h-screen max-w-5xl px-6 py-12">
      <p className="text-sm font-medium text-pink-400">Personal Workbench</p>
      <h1 className="mt-3 text-3xl font-semibold tracking-tight">个人工作台</h1>
      <p className="mt-4 max-w-2xl text-zinc-400">通过生成的 OpenAPI 类型加载并显示任务列表。</p>
      <TaskList />
    </main>
  );
}
