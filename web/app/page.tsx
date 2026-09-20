export default function Home() {
  return (
    <main className="mx-auto min-h-screen max-w-5xl px-6 py-12">
      <p className="text-sm font-medium text-pink-400">Personal Workbench</p>
      <h1 className="mt-3 text-3xl font-semibold tracking-tight">个人工作台</h1>
      <p className="mt-4 max-w-2xl text-zinc-400">
        Day 1 scaffold is ready. The task list will connect after the Go service and Gin handler pass their tests.
      </p>
      <section className="mt-10 rounded-xl border border-zinc-800 bg-zinc-950 p-6" aria-label="Task list placeholder">
        <div className="h-5 w-40 animate-pulse rounded bg-zinc-800 motion-reduce:animate-none" />
        <div className="mt-5 space-y-3">
          <div className="h-14 animate-pulse rounded-lg bg-zinc-900 motion-reduce:animate-none" />
          <div className="h-14 animate-pulse rounded-lg bg-zinc-900 motion-reduce:animate-none" />
        </div>
      </section>
    </main>
  );
}

