import { SignIn } from "@clerk/nextjs";
import { Logo } from "@/components/logo";
export default function SignInPage() {
  return (
    <main
      className="auth-layout"
      style={{
        minHeight: "100vh",
        display: "grid",
        gridTemplateColumns: "minmax(320px,.8fr) 1.2fr",
      }}
    >
      <section
        className="auth-intro"
        style={{
          padding: 40,
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          background: "#172033",
          color: "white",
        }}
      >
        <Logo />
        <div>
          <div className="eyebrow" style={{ color: "#a9baff" }}>
            Welcome back
          </div>
          <h1 className="title" style={{ margin: "16px 0" }}>
            Your career workspace is ready when you are.
          </h1>
          <p style={{ color: "#cbd5e1", lineHeight: 1.7 }}>
            Return to your profile, resumes, and next steps.
          </p>
        </div>
        <small style={{ color: "#94a3b8" }}>
          Secure identity provided by Clerk.
        </small>
      </section>
      <section style={{ display: "grid", placeItems: "center", padding: 28 }}>
        <SignIn
          routing="path"
          path="/sign-in"
          signUpUrl="/sign-up"
          forceRedirectUrl="/app"
        />
      </section>
    </main>
  );
}
