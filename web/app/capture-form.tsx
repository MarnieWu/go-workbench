"use client";

import { api } from "@/lib/api/client";
import { useState } from "react";

type SubmitState =
  | { kind: "idle" }
  | { kind: "submitting" }
  | { kind: "success"; captureId: string; requestId: string }
  | { kind: "http-error"; message: string; requestId?: string }
  | { kind: "network-error" };

export function CaptureForm() {
  const [inputText, setInputText] = useState("");
  const [submitState, setSubmitState] = useState<SubmitState>({ kind: "idle" });

  const isSubmitting = submitState.kind === "submitting";

  async function submitCapture() {
    const normalizedText = inputText.trim();
    if (normalizedText.length === 0 || isSubmitting) {
      return;
    }

    setSubmitState({ kind: "submitting" });

    try {
      const { data, error } = await api.POST("/v1/captures", {
        body: {
          idempotencyKey: crypto.randomUUID(),
          inputText: normalizedText,
          sourceType: "manual",
          excerpt: normalizedText,
        },
      });

      if (error) {
        setSubmitState({
          kind: "http-error",
          message: error.message,
          requestId: error.requestId,
        });
        return;
      }

      setSubmitState({
        kind: "success",
        captureId: data.id,
        requestId: data.requestId,
      });
      setInputText("");
    } catch {
      setSubmitState({ kind: "network-error" });
    }
  }

  return (
    <section className="mt-10 rounded-xl border border-zinc-800 bg-zinc-950 p-6" aria-label="Add to workbench">
      <h2 className="text-lg font-medium">加入工作台</h2>
      <div className="mt-4">
        <label htmlFor="capture-input" className="text-sm font-medium text-zinc-300">
          可见 TODO 内容
        </label>
        <textarea
          id="capture-input"
          value={inputText}
          onChange={(event) => {
            setInputText(event.target.value);
          }}
          rows={4}
          maxLength={250}
		  aria-describedby="capture-limit"
          className="mt-2 w-full resize-none rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 outline-none focus-visible:ring-2 focus-visible:ring-pink-300"
          placeholder="把当前可见内容中的任务加入工作台"
        />
		<p id="capture-limit" className="mt-2 text-xs text-zinc-500">
		  {inputText.length}/250，最多 250 个字符
		</p>
      </div>

      <button
        type="button"
        disabled={isSubmitting || inputText.trim().length === 0}
        onClick={() => {
          void submitCapture();
        }}
        className="mt-4 rounded-md bg-pink-400 px-4 py-2 text-sm font-medium text-zinc-950 outline-none hover:bg-pink-300 focus-visible:ring-2 focus-visible:ring-pink-200 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950 disabled:cursor-not-allowed disabled:bg-zinc-700 disabled:text-zinc-400"
      >
        {isSubmitting ? "提交中" : "加入"}
      </button>

      {submitState.kind === "success" ? (
        <p className="mt-3 text-sm text-zinc-300">
          已创建 Capture：{submitState.captureId} · Request ID: {submitState.requestId}
        </p>
      ) : null}

      {submitState.kind === "http-error" ? (
        <p className="mt-3 text-sm text-red-300" role="alert">
          提交失败：{submitState.message}
          {submitState.requestId ? ` · Request ID: ${submitState.requestId}` : ""}
        </p>
      ) : null}

      {submitState.kind === "network-error" ? (
        <p className="mt-3 text-sm text-red-300" role="alert">
          无法连接到 API，请确认服务正在运行。
        </p>
      ) : null}
    </section>
  );
}
