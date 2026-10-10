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
	const replaceTask = (updated: Task) => {
	  setState((current) => current.kind === "success"
	    ? { kind: "success", items: current.items.map((item) => item.id === updated.id ? updated : item) }
	    : current);
	};
    return (
      <section className="mt-10 rounded-xl border border-zinc-800 bg-zinc-950 p-6" aria-label="Task list">
        <h2 className="text-lg font-medium">任务</h2>
        <ul className="mt-5 space-y-3">
          {state.items.map((task) => (
			<TaskCard key={task.id} task={task} onUpdated={replaceTask} />
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

function TaskCard({ task, onUpdated }: { task: Task; onUpdated: (task: Task) => void }) {
  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState(task.title);
	const [description, setDescription] = useState(task.description ?? "");
  const [status, setStatus] = useState(task.status);
  const [priority, setPriority] = useState(task.priority);
  const [labels, setLabels] = useState(task.labels.join(", "));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<{ message: string; requestId?: string } | null>(null);

  async function save() {
    setSaving(true);
    setError(null);
    try {
      const { data, error: apiError } = await api.PATCH("/v1/tasks/{id}", {
        params: { path: { id: task.id } },
        body: {
          version: task.version,
		  projectId: task.projectId ?? null,
          title: title.trim(),
          description,
          status,
          priority,
          labels: labels.split(",").map((label) => label.trim()).filter(Boolean),
		  dueAt: task.dueAt ?? null,
        },
      });
      if (apiError) {
        setError({
          message: apiError.code === "VERSION_CONFLICT" ? "Task 已被其他请求更新。你的输入已保留，请刷新后重试。" : apiError.message,
          requestId: apiError.requestId,
        });
        return;
      }
      onUpdated(data);
      setEditing(false);
    } catch {
      setError({ message: "无法连接到 API。你的输入已保留。" });
    } finally {
      setSaving(false);
    }
  }

  if (!editing) {
    return (
      <li className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
        <p className="font-medium">{task.title}</p>
        <p className="mt-1 text-sm text-zinc-400">状态：{task.status} · 优先级：{task.priority} · v{task.version}</p>
        <button type="button" onClick={() => setEditing(true)} className="mt-3 rounded-md border border-zinc-700 px-3 py-1.5 text-sm focus-visible:ring-2 focus-visible:ring-pink-300">编辑</button>
      </li>
    );
  }

  return (
    <li className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
      <label className="block text-sm text-zinc-400" htmlFor={`task-title-${task.id}`}>标题</label>
      <input id={`task-title-${task.id}`} value={title} onChange={(event) => setTitle(event.target.value)} className="mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2" />
      <label className="mt-3 block text-sm text-zinc-400" htmlFor={`task-description-${task.id}`}>描述</label>
      <textarea id={`task-description-${task.id}`} value={description} onChange={(event) => setDescription(event.target.value)} rows={3} className="mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2" />
      <div className="mt-3 grid gap-3 sm:grid-cols-2">
        <label className="text-sm text-zinc-400">状态
          <select value={status} onChange={(event) => setStatus(event.target.value as Task["status"])} className="mt-1 block w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2">
            <option value="backlog">backlog</option><option value="in_progress">in_progress</option><option value="blocked">blocked</option><option value="done">done</option>
          </select>
        </label>
        <label className="text-sm text-zinc-400">优先级
          <select value={priority} onChange={(event) => setPriority(event.target.value as Task["priority"])} className="mt-1 block w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2">
            <option value="none">none</option><option value="low">low</option><option value="medium">medium</option><option value="high">high</option>
          </select>
        </label>
      </div>
      <label className="mt-3 block text-sm text-zinc-400" htmlFor={`task-labels-${task.id}`}>Labels（逗号分隔）</label>
      <input id={`task-labels-${task.id}`} value={labels} onChange={(event) => setLabels(event.target.value)} className="mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2" />
      <div className="mt-4 flex gap-3">
        <button type="button" disabled={saving || title.trim().length === 0} onClick={() => void save()} className="rounded-md bg-pink-400 px-4 py-2 text-sm font-medium text-zinc-950 disabled:bg-zinc-700">{saving ? "保存中" : "保存"}</button>
        <button type="button" disabled={saving} onClick={() => setEditing(false)} className="rounded-md border border-zinc-700 px-4 py-2 text-sm">取消</button>
      </div>
      {error ? <p className="mt-3 text-sm text-red-300" role="alert">{error.message}{error.requestId ? ` · Request ID: ${error.requestId}` : ""}</p> : null}
    </li>
  );
}
