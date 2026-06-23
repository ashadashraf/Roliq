import Link from "next/link";
import {
  ArrowRight,
  CheckCircle2,
  FileText,
  ShieldCheck,
  SlidersHorizontal,
  Sparkles,
} from "lucide-react";
import { Show, SignInButton, SignUpButton, UserButton } from "@clerk/nextjs";
import { Logo } from "@/components/logo";

export default function LandingPage() {
  return (
    <main>
      <header
        className="container landing-hero"
        style={{
          height: 80,
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <Logo />
        <nav
          aria-label="Primary"
          style={{ display: "flex", alignItems: "center", gap: 12 }}
        >
          <Show when="signed-out">
            <SignInButton>
              <button className="button ghost">Sign in</button>
            </SignInButton>
            <SignUpButton>
              <button className="button">Get started</button>
            </SignUpButton>
          </Show>
          <Show when="signed-in">
            <Link className="button secondary" href="/app">
              Open workspace
            </Link>
            <UserButton />
          </Show>
        </nav>
      </header>
      <section
        className="container"
        style={{
          padding: "88px 0 110px",
          display: "grid",
          gridTemplateColumns: "minmax(0,1.25fr) minmax(320px,.75fr)",
          gap: 64,
          alignItems: "center",
        }}
      >
        <div>
          <div className="eyebrow">
            A calmer way to move your career forward
          </div>
          <h1
            className="display"
            style={{ maxWidth: 840, margin: "22px 0 28px" }}
          >
            Your experience deserves a better system.
          </h1>
          <p
            className="muted"
            style={{ fontSize: "1.2rem", lineHeight: 1.7, maxWidth: 680 }}
          >
            Roliq turns your career history into a clear, reusable
            profile—giving you a trustworthy foundation for more relevant
            applications, while you remain in control.
          </p>
          <div
            style={{
              display: "flex",
              gap: 12,
              flexWrap: "wrap",
              marginTop: 34,
            }}
          >
            <Show when="signed-out">
              <SignUpButton>
                <button className="button">
                  Build your career profile <ArrowRight size={17} />
                </button>
              </SignUpButton>
            </Show>
            <Show when="signed-in">
              <Link className="button" href="/app">
                Continue to Roliq <ArrowRight size={17} />
              </Link>
            </Show>
            <a className="button secondary" href="#how-it-works">
              See how it works
            </a>
          </div>
        </div>
        <div
          className="card"
          style={{ padding: 26, transform: "rotate(1deg)" }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div>
              <div className="muted" style={{ fontSize: 13 }}>
                Career profile
              </div>
              <strong style={{ fontSize: 20 }}>Your professional story</strong>
            </div>
            <span className="pill">
              <Sparkles size={14} /> Organized
            </span>
          </div>
          <div style={{ margin: "28px 0 12px", display: "grid", gap: 11 }}>
            <div className="progress">
              <span style={{ width: "72%" }} />
            </div>
            <small className="muted">
              A single source of truth for your experience, education and
              skills.
            </small>
          </div>
          {[
            "Experience structured",
            "Skills normalized",
            "Resume safely stored",
          ].map((item) => (
            <div
              key={item}
              style={{
                display: "flex",
                gap: 10,
                padding: "13px 0",
                borderTop: "1px solid var(--line)",
              }}
            >
              <CheckCircle2 size={19} color="#18794e" />
              <span>{item}</span>
            </div>
          ))}
        </div>
      </section>
      <section
        id="how-it-works"
        style={{ background: "#172033", color: "white", padding: "110px 0" }}
      >
        <div className="container">
          <div className="eyebrow" style={{ color: "#a9baff" }}>
            Designed around trust
          </div>
          <h2
            className="title"
            style={{ maxWidth: 680, margin: "18px 0 50px" }}
          >
            One thoughtful foundation for every opportunity.
          </h2>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit,minmax(230px,1fr))",
              gap: 18,
            }}
          >
            {[
              [
                FileText,
                "Bring your history",
                "Upload a PDF or DOCX, or build your profile manually from the start.",
              ],
              [
                SlidersHorizontal,
                "Keep it accurate",
                "Review and update every detail. Your career record stays editable and yours.",
              ],
              [
                ShieldCheck,
                "Stay in control",
                "Roliq is designed around explicit choices, tenant isolation and auditable actions.",
              ],
            ].map(([Icon, title, copy]) => (
              <article
                key={String(title)}
                style={{
                  border: "1px solid #344054",
                  borderRadius: 18,
                  padding: 25,
                  background: "#1d2939",
                }}
              >
                <Icon size={25} color="#a9baff" />
                <h3 style={{ fontSize: 19, margin: "26px 0 10px" }}>
                  {String(title)}
                </h3>
                <p style={{ color: "#cbd5e1", lineHeight: 1.65, margin: 0 }}>
                  {String(copy)}
                </p>
              </article>
            ))}
          </div>
        </div>
      </section>
      <section
        className="container"
        style={{ padding: "110px 0", textAlign: "center" }}
      >
        <div className="eyebrow">Begin with what you already have</div>
        <h2 className="title" style={{ margin: "18px auto", maxWidth: 760 }}>
          Make your career information useful, consistent and ready for what
          comes next.
        </h2>
        <p
          className="muted"
          style={{ maxWidth: 620, margin: "0 auto 30px", lineHeight: 1.7 }}
        >
          No invented scores. No opaque automation. Just a durable professional
          profile built from your real experience.
        </p>
        <Show when="signed-out">
          <SignUpButton>
            <button className="button">Create your Roliq workspace</button>
          </SignUpButton>
        </Show>
      </section>
      <footer style={{ borderTop: "1px solid var(--line)", padding: "28px 0" }}>
        <div
          className="container"
          style={{
            display: "flex",
            justifyContent: "space-between",
            gap: 20,
            flexWrap: "wrap",
          }}
        >
          <Logo />
          <span className="muted">
            © {new Date().getFullYear()} Roliq. Your career data, handled with
            care.
          </span>
        </div>
      </footer>
    </main>
  );
}
