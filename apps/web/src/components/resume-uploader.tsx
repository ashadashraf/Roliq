"use client";

import { useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { FileCheck2, LoaderCircle, UploadCloud } from "lucide-react";
import { useRoliqClient } from "@/lib/api";

const accepted = [
  "application/pdf",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
];
async function sha256(file: File) {
  const bytes = await file.arrayBuffer();
  const hash = await crypto.subtle.digest("SHA-256", bytes);
  return Array.from(new Uint8Array(hash))
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");
}
export function ResumeUploader({ onReady }: { onReady?: () => void }) {
  const api = useRoliqClient();
  const client = useQueryClient();
  const inputRef = useRef<HTMLInputElement>(null);
  const [state, setState] = useState<{
    busy: boolean;
    progress: string;
    error: string;
  }>({ busy: false, progress: "", error: "" });
  const query = useQuery({
    queryKey: ["resumes"],
    queryFn: api.listResumes,
    refetchInterval: (q) =>
      q.state.data?.items.some((r) =>
        ["uploaded", "scanning"].includes(r.status),
      )
        ? 3500
        : false,
  });
  async function upload(file: File) {
    if (!accepted.includes(file.type)) {
      setState({
        busy: false,
        progress: "",
        error: "Choose a PDF or DOCX file.",
      });
      return;
    }
    if (file.size > 10 * 1024 * 1024) {
      setState({
        busy: false,
        progress: "",
        error: "Files must be 10 MB or smaller.",
      });
      return;
    }
    setState({
      busy: true,
      progress: "Calculating secure checksum…",
      error: "",
    });
    try {
      const checksum = await sha256(file);
      setState((s) => ({ ...s, progress: "Creating secure upload…" }));
      const intent = await api.createUpload({
        fileName: file.name,
        contentType: file.type,
        sizeBytes: file.size,
        checksumSha256: checksum,
      });
      setState((s) => ({ ...s, progress: "Uploading to quarantine storage…" }));
      const response = await fetch(intent.uploadUrl, {
        method: "PUT",
        body: file,
        headers: intent.requiredHeaders,
      });
      if (!response.ok)
        throw new Error("The storage service rejected the upload.");
      setState((s) => ({ ...s, progress: "Verifying uploaded file…" }));
      await api.completeUpload(intent.uploadId);
      await Promise.all([
        client.invalidateQueries({ queryKey: ["resumes"] }),
        client.invalidateQueries({ queryKey: ["dashboard"] }),
      ]);
      setState({
        busy: false,
        progress: "Upload verified. Security scanning has started.",
        error: "",
      });
      onReady?.();
    } catch (error) {
      setState({
        busy: false,
        progress: "",
        error:
          error instanceof Error
            ? error.message
            : "The resume could not be uploaded.",
      });
    }
  }
  return (
    <div>
      <div
        className="card"
        style={{ padding: 28, borderStyle: "dashed", textAlign: "center" }}
      >
        <UploadCloud
          size={34}
          color="#3157d5"
          style={{ margin: "0 auto 14px" }}
        />
        <h3 style={{ margin: "0 0 8px" }}>Upload your original resume</h3>
        <p
          className="muted"
          style={{ margin: "0 auto 20px", maxWidth: 480, lineHeight: 1.6 }}
        >
          PDF or DOCX, up to 10 MB. Files are quarantined, verified, and scanned
          before they become available.
        </p>
        <input
          ref={inputRef}
          hidden
          type="file"
          accept=".pdf,.docx,application/pdf,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
          onChange={(event) => {
            const file = event.target.files?.[0];
            if (file) upload(file);
            event.target.value = "";
          }}
        />
        <button
          className="button"
          disabled={state.busy}
          onClick={() => inputRef.current?.click()}
        >
          {state.busy ? <LoaderCircle size={17} /> : <UploadCloud size={17} />}{" "}
          {state.busy ? "Working…" : "Choose resume"}
        </button>
        {state.progress && (
          <p role="status" style={{ color: "var(--success)", marginBottom: 0 }}>
            {state.progress}
          </p>
        )}
        {state.error && (
          <p role="alert" className="field-error" style={{ marginBottom: 0 }}>
            {state.error}
          </p>
        )}
      </div>
      {query.data?.items.length ? (
        <div className="card" style={{ marginTop: 20, overflow: "hidden" }}>
          {query.data.items.map((resume, index) => (
            <div
              key={resume.id}
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                gap: 16,
                padding: 18,
                borderTop: index ? "1px solid var(--line)" : "none",
              }}
            >
              <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                <FileCheck2
                  color={resume.status === "ready" ? "#18794e" : "#3157d5"}
                />
                <div>
                  <strong>{resume.fileName}</strong>
                  <div className="muted" style={{ fontSize: 12, marginTop: 4 }}>
                    {new Date(resume.createdAt).toLocaleString()}
                  </div>
                </div>
              </div>
              <span className="pill">{resume.status}</span>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}
