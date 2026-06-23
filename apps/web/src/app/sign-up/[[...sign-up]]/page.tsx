import { SignUp } from "@clerk/nextjs";
import { Logo } from "@/components/logo";
export default function SignUpPage() {
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
            Start clearly
          </div>
          <h1 className="title" style={{ margin: "16px 0" }}>
            Give your professional story a reliable home.
          </h1>
          <p style={{ color: "#cbd5e1", lineHeight: 1.7 }}>
            A private workspace is created for you automatically.
          </p>
        </div>
        <small style={{ color: "#94a3b8" }}>
          By continuing, you agree to handle submitted information truthfully.
        </small>
      </section>
      <section style={{ display: "grid", placeItems: "center", padding: 28 }}>
        <SignUp
          routing="path"
          path="/sign-up"
          signInUrl="/sign-in"
          forceRedirectUrl="/onboarding"
        />
      </section>
    </main>
  );
}
