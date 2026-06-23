import type { Metadata } from "next";
import { ProfileEditor } from "@/components/profile-editor";
export const metadata: Metadata = { title: "Career profile" };
export default function ProfilePage() {
  return (
    <>
      <div className="eyebrow">Source of truth</div>
      <h1
        className="title"
        style={{ fontSize: "clamp(2rem,4vw,3rem)", margin: "10px 0" }}
      >
        Career profile
      </h1>
      <p
        className="muted"
        style={{ maxWidth: 700, lineHeight: 1.7, marginBottom: 30 }}
      >
        Keep your professional history factual and current. This structured
        record will become the foundation for later role-specific documents.
      </p>
      <ProfileEditor />
    </>
  );
}
