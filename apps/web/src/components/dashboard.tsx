"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowRight,
  CheckCircle2,
  Clock3,
  FileText,
  UserRound,
} from "lucide-react";
import { useRoliqClient } from "@/lib/api";
import { resumeStatusLabel } from "@/lib/resume-status";

export function Dashboard() {
  const api = useRoliqClient();
  const { data, isLoading, error } = useQuery({
    queryKey: ["dashboard"],
    queryFn: api.dashboard,
    refetchInterval: (q) =>
      q.state.data?.resumes.some((r) =>
        ["uploaded", "scanning"].includes(r.status),
      )
        ? 4000
        : false,
  });
  if (isLoading) return <p className="muted">Loading your workspace…</p>;
  if (error || !data)
    return (
      <p role="alert" className="field-error">
        Your dashboard could not be loaded.
      </p>
    );
  const complete = data.onboarding.status === "completed";
  return (
    <>
      <div
        style={{
          display: "flex",
          alignItems: "end",
          justifyContent: "space-between",
          gap: 20,
          flexWrap: "wrap",
        }}
      >
        <div>
          <div className="eyebrow">Overview</div>
          <h1
            className="title"
            style={{ fontSize: "clamp(2.1rem,4vw,3.2rem)", margin: "10px 0" }}
          >
            Your career, in one clear view.
          </h1>
          <p className="muted">
            Keep your core information accurate before using it anywhere else.
          </p>
        </div>
        {!complete && (
          <Link href="/onboarding" className="button">
            Continue setup <ArrowRight size={17} />
          </Link>
        )}
      </div>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit,minmax(270px,1fr))",
          gap: 18,
          marginTop: 34,
        }}
      >
        <section className="card" style={{ padding: 24 }}>
          <div style={{ display: "flex", justifyContent: "space-between" }}>
            <span className="muted">Profile strength</span>
            <UserRound size={20} color="#3157d5" />
          </div>
          <strong
            style={{
              display: "block",
              fontFamily: "Georgia",
              fontSize: 44,
              margin: "20px 0 14px",
            }}
          >
            {data.profileCompletion}%
          </strong>
          <div className="progress">
            <span style={{ width: `${data.profileCompletion}%` }} />
          </div>
          <Link
            href="/app/profile"
            style={{
              display: "inline-flex",
              gap: 6,
              alignItems: "center",
              color: "var(--brand)",
              fontWeight: 700,
              marginTop: 22,
            }}
          >
            Review profile <ArrowRight size={15} />
          </Link>
        </section>
        <section className="card" style={{ padding: 24 }}>
          <div style={{ display: "flex", justifyContent: "space-between" }}>
            <span className="muted">Resumes</span>
            <FileText size={20} color="#3157d5" />
          </div>
          <strong
            style={{
              display: "block",
              fontFamily: "Georgia",
              fontSize: 44,
              margin: "20px 0 6px",
            }}
          >
            {data.resumes.length}
          </strong>
          <p className="muted" style={{ margin: "0 0 24px" }}>
            Securely submitted documents
          </p>
          <Link
            href="/app/resumes"
            style={{
              display: "inline-flex",
              gap: 6,
              alignItems: "center",
              color: "var(--brand)",
              fontWeight: 700,
            }}
          >
            Manage resumes <ArrowRight size={15} />
          </Link>
        </section>
        <section className="card" style={{ padding: 24 }}>
          <div style={{ display: "flex", justifyContent: "space-between" }}>
            <span className="muted">Setup status</span>
            {complete ? (
              <CheckCircle2 size={20} color="#18794e" />
            ) : (
              <Clock3 size={20} color="#b54708" />
            )}
          </div>
          <strong
            style={{
              display: "block",
              fontFamily: "Georgia",
              fontSize: 30,
              margin: "24px 0 8px",
            }}
          >
            {complete ? "Complete" : "In progress"}
          </strong>
          <p className="muted" style={{ lineHeight: 1.6 }}>
            Step {data.onboarding.currentStep} of 4. Your progress is saved
            automatically.
          </p>
        </section>
      </div>
      {data.resumes.length > 0 && (
        <section style={{ marginTop: 36 }}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <h2 style={{ fontFamily: "Georgia", fontSize: 27 }}>
              Recent resumes
            </h2>
            <Link href="/app/resumes" className="muted">
              View all
            </Link>
          </div>
          <div className="card" style={{ overflow: "hidden" }}>
            {data.resumes.slice(0, 3).map((resume, index) => (
              <div
                key={resume.id}
                style={{
                  padding: 20,
                  display: "flex",
                  justifyContent: "space-between",
                  gap: 16,
                  borderTop: index ? "1px solid var(--line)" : "none",
                }}
              >
                <div>
                  <strong>{resume.fileName}</strong>
                  <div className="muted" style={{ fontSize: 13, marginTop: 5 }}>
                    {new Date(resume.createdAt).toLocaleDateString()}
                  </div>
                </div>
                <span className="pill">{resumeStatusLabel(resume.status)}</span>
              </div>
            ))}
          </div>
        </section>
      )}
    </>
  );
}
