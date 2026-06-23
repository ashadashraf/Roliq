"use client";

import { useQuery } from "@tanstack/react-query";
import { ApiError } from "@roliq/api-client";
import { useRoliqClient } from "@/lib/api";

export function BootstrapGate({ children }: { children: React.ReactNode }) {
  const api = useRoliqClient();
  const query = useQuery({
    queryKey: ["session"],
    queryFn: api.bootstrap,
    retry: (count, error) =>
      !(error instanceof ApiError && error.status < 500) && count < 2,
  });
  if (query.isLoading)
    return (
      <main
        style={{ minHeight: "100vh", display: "grid", placeItems: "center" }}
      >
        <div role="status" className="muted">
          Preparing your private workspace…
        </div>
      </main>
    );
  if (query.isError)
    return (
      <main
        style={{
          minHeight: "100vh",
          display: "grid",
          placeItems: "center",
          padding: 24,
        }}
      >
        <div className="card" style={{ maxWidth: 520, padding: 30 }}>
          <div className="eyebrow">Workspace unavailable</div>
          <h1 style={{ fontFamily: "Georgia", fontSize: 32 }}>
            We couldn’t open Roliq.
          </h1>
          <p className="muted">
            {query.error instanceof Error
              ? query.error.message
              : "Please try again."}
          </p>
          <button className="button" onClick={() => query.refetch()}>
            Try again
          </button>
        </div>
      </main>
    );
  return children;
}
