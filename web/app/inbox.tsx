"use client";

import { api } from "@/lib/api/client";
import type { components } from "@/lib/api/generated/schema";
import { useEffect, useState } from "react";

type Candidate = components["schemas"]["Candidate"];
type InboxState =
  | { kind: "loading" }
  | { kind: "empty" }
  | { kind: "success"; items: Candidate[] }
  | { kind: "http-error"; requestId?: string }
  | { kind: "network-error" };

async function requestInbox(): Promise<InboxState> {
  try {
    const { data, error } = await api.GET("/v1/inbox");
    if (error) {
      return { kind: "http-error", requestId: error.requestId };
    }
    return data.items.length === 0 ? { kind: "empty" } : { kind: "success", items: data.items };
  } catch {
    return { kind: "network-error" };
  }
}

export function Inbox() {
  const [state, setState] = useState<InboxState>({ kind: "loading" });

  useEffect(() => {
    let cancelled = false;
    void requestInbox().then((nextState) => {
      if (!cancelled) {
        setState(nextState);
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);

  async function retry() {
    setState({ kind: "loading" });
    setState(await requestInbox());
  }

  function removeCandidate(candidateId: string) {
    setState((current) => {
      if (current.kind !== "success") {
        return current;
      }
      const items = current.items.filter((item) => item.id !== candidateId);
      return items.length === 0 ? { kind: "empty" } : { kind: "success", items };
    });
  }

  if (state.kind === "loading") {
    return (
      <section className="mt-10 rounded-xl border border-zinc-800 bg-zinc-950 p-6" aria-busy="true" aria-label="Inbox">
        <p className="sr-only">正在加载待审核建议</p>
        <div className="h-5 w-32 animate-pulse rounded bg-zinc-800 motion-reduce:animate-none" />
        <div className="mt-5 h-28 animate-pulse rounded-lg bg-zinc-900 motion-reduce:animate-none" />
      </section>
    );
  }

  if (state.kind === "empty") {
    return (
      <section className="mt-10 rounded-xl border border-zinc-800 bg-zinc-950 p-6" aria-label="Inbox">
        <h2 className="text-lg font-medium">Inbox</h2>
        <p className="mt-2 text-sm text-zinc-400">还没有待审核建议。提交 Capture 后，Worker 生成的建议会显示在这里。</p>
      </section>
    );
  }

  if (state.kind === "success") {
    return (
      <section className="mt-10 rounded-xl border border-zinc-800 bg-zinc-950 p-6" aria-label="Inbox">
        <h2 className="text-lg font-medium">Inbox</h2>
        <ul className="mt-5 space-y-3">
          {state.items.map((candidate) => (
            <CandidateCard key={candidate.id} candidate={candidate} onResolved={removeCandidate} />
          ))}
        </ul>
      </section>
    );
  }

  return (
    <section className="mt-10 rounded-xl border border-red-900/60 bg-zinc-950 p-6" aria-label="Inbox" role="alert">
      <h2 className="text-lg font-medium">无法加载 Inbox</h2>
      <p className="mt-2 text-sm text-zinc-400">
        {state.kind === "network-error" ? "无法连接到 API。" : "服务暂时无法处理请求。"}
      </p>
      {state.kind === "http-error" && state.requestId ? <p className="mt-2 text-sm text-zinc-500">Request ID: {state.requestId}</p> : null}
      <button
        type="button"
        className="mt-4 rounded-md bg-pink-400 px-4 py-2 text-sm font-medium text-zinc-950 outline-none focus-visible:ring-2 focus-visible:ring-pink-200 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950"
        onClick={() => void retry()}
      >
        重试
      </button>
    </section>
  );
}

type MutationState = { kind: "idle" } | { kind: "submitting" } | { kind: "error"; message: string; requestId?: string };

function CandidateCard({ candidate, onResolved }: { candidate: Candidate; onResolved: (candidateId: string) => void }) {
  const [title, setTitle] = useState(candidate.proposedTitle);
  const [description, setDescription] = useState(candidate.proposedDescription ?? "");
  const [labels, setLabels] = useState(candidate.labels.join(", "));
  const [mutation, setMutation] = useState<MutationState>({ kind: "idle" });
  const submitting = mutation.kind === "submitting";

  async function reject() {
    setMutation({ kind: "submitting" });
    try {
      const { error } = await api.POST("/v1/candidates/{id}/reject", { params: { path: { id: candidate.id } } });
      if (error) {
        setMutation({ kind: "error", message: error.message, requestId: error.requestId });
        return;
      }
      onResolved(candidate.id);
    } catch {
      setMutation({ kind: "error", message: "无法连接到 API" });
    }
  }

  async function accept() {
    if (title.trim().length === 0) {
      setMutation({ kind: "error", message: "标题不能为空" });
      return;
    }
    setMutation({ kind: "submitting" });
    try {
      const { error } = await api.POST("/v1/candidates/{id}/accept", {
        params: { path: { id: candidate.id } },
        body: {
          title: title.trim(),
          description: description.trim() || null,
          projectId: candidate.proposedProjectId,
          labels: labels.split(",").map((label) => label.trim()).filter(Boolean),
        },
      });
      if (error) {
        const message = error.code === "CANDIDATE_STATE_CONFLICT" ? "Candidate 状态已变化。请刷新后确认。" : error.message;
        setMutation({ kind: "error", message, requestId: error.requestId });
        return;
      }
      onResolved(candidate.id);
    } catch {
      setMutation({ kind: "error", message: "无法连接到 API" });
    }
  }

  return (
    <li className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
      <label className="block text-sm text-zinc-400" htmlFor={`candidate-title-${candidate.id}`}>标题</label>
      <input
        id={`candidate-title-${candidate.id}`}
        value={title}
        maxLength={200}
        onChange={(event) => setTitle(event.target.value)}
        className="mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 outline-none focus-visible:ring-2 focus-visible:ring-pink-300"
      />
      <label className="mt-3 block text-sm text-zinc-400" htmlFor={`candidate-description-${candidate.id}`}>描述</label>
      <textarea
        id={`candidate-description-${candidate.id}`}
        value={description}
        onChange={(event) => setDescription(event.target.value)}
        rows={3}
        className="mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 outline-none focus-visible:ring-2 focus-visible:ring-pink-300"
      />
      <label className="mt-3 block text-sm text-zinc-400" htmlFor={`candidate-labels-${candidate.id}`}>Labels（逗号分隔）</label>
      <input
        id={`candidate-labels-${candidate.id}`}
        value={labels}
        onChange={(event) => setLabels(event.target.value)}
        className="mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 outline-none focus-visible:ring-2 focus-visible:ring-pink-300"
      />
      <div className="mt-4 flex flex-wrap gap-3">
        <button type="button" disabled={submitting} onClick={() => void accept()} className="rounded-md bg-pink-400 px-4 py-2 text-sm font-medium text-zinc-950 disabled:bg-zinc-700">
          {submitting ? "处理中" : "接受"}
        </button>
        <button type="button" disabled={submitting} onClick={() => void reject()} className="rounded-md border border-zinc-700 px-4 py-2 text-sm disabled:text-zinc-500">
          拒绝
        </button>
      </div>
      {mutation.kind === "error" ? (
        <p className="mt-3 text-sm text-red-300" role="alert">
          {mutation.message}{mutation.requestId ? ` · Request ID: ${mutation.requestId}` : ""}
        </p>
      ) : null}
    </li>
  );
}
