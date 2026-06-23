"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  ArrowRight,
  FileText,
  Keyboard,
  PartyPopper,
} from "lucide-react";
import { CareerProfile } from "@roliq/api-client";
import { useRoliqClient } from "@/lib/api";
import { ProfileEditor } from "./profile-editor";
import { ResumeUploader } from "./resume-uploader";

export function OnboardingFlow() {
  const api = useRoliqClient();
  const router = useRouter();
  const queryClient = useQueryClient();
  const session = useQuery({ queryKey: ["session"], queryFn: api.bootstrap });
  const profileQuery = useQuery({
    queryKey: ["profile"],
    queryFn: api.getProfile,
    enabled: !!session.data,
  });
  const [step, setStep] = useState(1);
  const [basics, setBasics] = useState<CareerProfile | null>(null);
  const currentMethod = session.data?.onboarding.profileMethod;
  const resumes = useQuery({
    queryKey: ["resumes"],
    queryFn: api.listResumes,
    enabled: currentMethod === "resume" && step === 3,
  });
  useEffect(() => {
    // Editable onboarding state is initialized from the persisted server checkpoint.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (session.data) setStep(session.data.onboarding.currentStep);
  }, [session.data]);
  useEffect(() => {
    // The form keeps an isolated draft after its initial persisted value arrives.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (profileQuery.data) setBasics(profileQuery.data);
  }, [profileQuery.data]);
  const progress = useMutation({
    mutationFn: api.updateOnboarding,
    onSuccess: (data) => {
      setStep(data.currentStep);
      queryClient.invalidateQueries({ queryKey: ["session"] });
    },
  });
  async function saveBasics() {
    if (!basics) return;
    await api.saveProfile(basics);
    progress.mutate({ currentStep: 2, status: "in_progress" });
  }
  async function choose(method: "resume" | "manual") {
    await api.updateOnboarding({
      profileMethod: method,
      currentStep: 3,
      status: "in_progress",
    });
    await queryClient.invalidateQueries({ queryKey: ["session"] });
    setStep(3);
  }
  async function finish() {
    await api.updateOnboarding({ currentStep: 4, status: "completed" });
    await queryClient.invalidateQueries();
    router.push("/app");
  }
  if (session.isLoading || !basics)
    return <p className="muted">Preparing onboarding…</p>;
  const method = currentMethod;
  return (
    <div
      className="container"
      style={{ maxWidth: 920, padding: "34px 0 80px" }}
    >
      <header
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: 34,
        }}
      >
        <Link href="/" style={{ fontWeight: 800, fontSize: 20 }}>
          Roliq
        </Link>
        <span className="muted">Step {step} of 4</span>
      </header>
      <div className="progress" aria-label={`Step ${step} of 4`}>
        <span style={{ width: `${step * 25}%` }} />
      </div>
      <div style={{ margin: "38px 0 28px" }}>
        {step === 1 && (
          <>
            <div className="eyebrow">Let’s begin with the basics</div>
            <h1
              className="title"
              style={{ fontSize: "clamp(2.2rem,5vw,3.6rem)", margin: "12px 0" }}
            >
              How should your professional story begin?
            </h1>
            <p className="muted">
              Start with your current professional identity. You can revise
              every detail later.
            </p>
          </>
        )}
        {step === 2 && (
          <>
            <div className="eyebrow">Choose your starting point</div>
            <h1
              className="title"
              style={{ fontSize: "clamp(2.2rem,5vw,3.6rem)", margin: "12px 0" }}
            >
              Bring a resume, or build from scratch.
            </h1>
            <p className="muted">
              Both paths create the same private, structured career workspace.
            </p>
          </>
        )}
        {step === 3 && (
          <>
            <div className="eyebrow">Build your foundation</div>
            <h1
              className="title"
              style={{ fontSize: "clamp(2.2rem,5vw,3.6rem)", margin: "12px 0" }}
            >
              {method === "resume"
                ? "Securely add your source resume."
                : "Capture the experience that matters."}
            </h1>
          </>
        )}
        {step === 4 && (
          <>
            <div className="eyebrow">Ready for what comes next</div>
            <h1
              className="title"
              style={{ fontSize: "clamp(2.2rem,5vw,3.6rem)", margin: "12px 0" }}
            >
              Your Roliq workspace is ready.
            </h1>
          </>
        )}
      </div>
      {step === 1 && (
        <section
          className="card"
          style={{ padding: 28, display: "grid", gap: 16 }}
        >
          <label className="label">
            Professional headline
            <input
              className="input"
              maxLength={160}
              value={basics.headline}
              placeholder="Product designer, platform engineer, operations leader…"
              onChange={(e) =>
                setBasics({ ...basics, headline: e.target.value })
              }
            />
          </label>
          <div
            className="responsive-grid"
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(3,1fr)",
              gap: 14,
            }}
          >
            <label className="label">
              City
              <input
                className="input"
                value={basics.city}
                onChange={(e) => setBasics({ ...basics, city: e.target.value })}
              />
            </label>
            <label className="label">
              Country code
              <input
                className="input"
                maxLength={2}
                value={basics.countryCode}
                placeholder="IN"
                onChange={(e) =>
                  setBasics({
                    ...basics,
                    countryCode: e.target.value.toUpperCase(),
                  })
                }
              />
            </label>
            <label className="label">
              Time zone
              <input
                className="input"
                value={basics.timeZone}
                onChange={(e) =>
                  setBasics({ ...basics, timeZone: e.target.value })
                }
              />
            </label>
          </div>
          <div style={{ display: "flex", justifyContent: "flex-end" }}>
            <button
              className="button"
              onClick={saveBasics}
              disabled={!basics.headline || progress.isPending}
            >
              Continue <ArrowRight size={17} />
            </button>
          </div>
        </section>
      )}
      {step === 2 && (
        <div
          className="responsive-grid"
          style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 18 }}
        >
          <button
            className="card"
            style={{ padding: 28, textAlign: "left", color: "inherit" }}
            onClick={() => choose("resume")}
          >
            <FileText size={29} color="#3157d5" />
            <h2 style={{ fontFamily: "Georgia", fontSize: 25 }}>
              Upload a resume
            </h2>
            <p className="muted" style={{ lineHeight: 1.6 }}>
              Start from a PDF or DOCX. Roliq securely stores and scans it
              before any later processing is allowed.
            </p>
          </button>
          <button
            className="card"
            style={{ padding: 28, textAlign: "left", color: "inherit" }}
            onClick={() => choose("manual")}
          >
            <Keyboard size={29} color="#3157d5" />
            <h2 style={{ fontFamily: "Georgia", fontSize: 25 }}>
              Enter details manually
            </h2>
            <p className="muted" style={{ lineHeight: 1.6 }}>
              Build a complete, structured record of your skills, experience and
              education.
            </p>
          </button>
        </div>
      )}
      {step === 3 && method === "resume" && (
        <>
          <ResumeUploader />
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              marginTop: 24,
            }}
          >
            <button
              className="button secondary"
              onClick={() => progress.mutate({ currentStep: 2 })}
            >
              <ArrowLeft size={16} /> Back
            </button>
            <button
              className="button"
              disabled={!resumes.data?.items.length}
              onClick={() => progress.mutate({ currentStep: 4 })}
            >
              Review setup <ArrowRight size={16} />
            </button>
          </div>
        </>
      )}
      {step === 3 && method === "manual" && (
        <>
          <ProfileEditor
            compact
            onSaved={() => progress.mutate({ currentStep: 4 })}
          />
          <button
            className="button secondary"
            onClick={() => progress.mutate({ currentStep: 2 })}
          >
            <ArrowLeft size={16} /> Back
          </button>
        </>
      )}
      {step === 4 && (
        <section className="card" style={{ padding: 34, textAlign: "center" }}>
          <PartyPopper
            size={38}
            color="#3157d5"
            style={{ margin: "0 auto 15px" }}
          />
          <h2 style={{ fontFamily: "Georgia", fontSize: 30, margin: "8px 0" }}>
            A reliable source of truth
          </h2>
          <p
            className="muted"
            style={{ maxWidth: 560, margin: "0 auto 25px", lineHeight: 1.7 }}
          >
            Your personal workspace, career information, and submitted resumes
            are stored in your organization-scoped account. Continue to the
            dashboard to review everything.
          </p>
          <button className="button" onClick={finish}>
            Open my workspace <ArrowRight size={17} />
          </button>
        </section>
      )}
    </div>
  );
}
