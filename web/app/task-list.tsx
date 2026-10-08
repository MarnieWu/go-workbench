"use client";

import { api } from "@/lib/api/client";
import type { components } from "@/lib/api/generated/schema";
import { useEffect, useState } from "react";

type Task = components["schemas"]["Task"];

type State =
  | { kind: "loading" }
  | { kind: "success"; items: Task[] }
  | { kind: "empty" }
  | { kind: "http-error"; message: string; requestId?: string }
  | { kind: "network-error" };

async function requestTasks(): Promise<State> {
  try {
    const { data, error } = await api.GET("/v1/tasks", {
      params: {
        query: {
          status: "backlog",
        },
      },
    });

    console.log({ data, error });

    if (error) {
      return {
        kind: "http-error",
        message: error.message,
        requestId: error.requestId,
      };
    }

    return data.items.length === 0
      ? { kind: "empty" }
      : {
          kind: "success",
          items: data.items,
        };
  } catch {
    return { kind: "network-error" };
  }
}

export function TaskList() {
  const [state, setState] = useState<State>({ kind: "loading" });

  useEffect(() => {
    let cancelled = false;

    void requestTasks().then((nextState) => {
      if (!cancelled) {
        setState(nextState);
      }
    });

    return () => {
      cancelled = true;
    };
  }, []);

  const retry = async () => {
    setState({ kind: "loading" });
    setState(await requestTasks());
  };

  if (state.kind === "loading") {
    return (
      <section
        className="mt-10 rounded-xl border border-zinc-800 bg-zinc-950 p-6"
        aria-busy="true"
        aria-label="Task list"
      >
        <p className="sr-only">正在加载任务</p>
        <div className="h-5 w-40 animate-pulse rounded bg-zinc-800 motion-reduce:animate-none" />
        <div className="mt-5 space-y-3">
          <div className="h-14 animate-pulse rounded-lg bg-zinc-900 motion-reduce:animate-none" />
          <div className="h-14 animate-pulse rounded-lg bg-zinc-900 motion-reduce:animate-none" />
        </div>
      </section>
    );
  }

  if (state.kind === "empty") {
    return (
      <section className="mt-10 rounded-xl border border-zinc-800 bg-zinc-950 p-6" aria-label="Task list">
        <h2 className="text-lg font-medium">暂无任务</h2>
        <p className="mt-2 text-sm text-zinc-400">当前页面仅支持查看任务。</p>
      </section>
    );
  }

  if (state.kind === "success") {
    return (
      <section className="mt-10 rounded-xl border border-zinc-800 bg-zinc-950 p-6" aria-label="Task list">
        <h2 className="text-lg font-medium">任务</h2>
        <ul className="mt-5 space-y-3">
          {state.items.map((task) => (
            <li key={task.id} className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
              <p className="font-medium">{task.title}</p>
              <p className="mt-1 text-sm text-zinc-400">
                状态：{task.status} · 优先级：{task.priority}
              </p>
            </li>
          ))}
        </ul>
      </section>
    );
  }

  const isHttpError = state.kind === "http-error";

  return (
    <section className="mt-10 rounded-xl border border-red-900/60 bg-zinc-950 p-6" aria-label="Task list" role="alert">
      <h2 className="text-lg font-medium">无法加载任务</h2>
      <p className="mt-2 text-sm text-zinc-400">
        {isHttpError ? "服务暂时无法处理请求，请重试。" : "无法连接到 API，请确认服务正在运行后重试。"}
      </p>

      <p className="mt-2 text-sm wrap-break-word text-zinc-500">
        {isHttpError && state.requestId ? `Request ID: ${state.requestId}` : "无 request ID"}
      </p>

      <button
        type="button"
        className="mt-4 rounded-md bg-pink-400 px-4 py-2 text-sm font-medium text-zinc-950 outline-none hover:bg-pink-300 focus-visible:ring-2 focus-visible:ring-pink-200 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950"
        onClick={() => {
          void retry();
        }}
      >
        重试
      </button>
    </section>
  );
}
