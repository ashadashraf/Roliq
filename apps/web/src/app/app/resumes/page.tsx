import type { Metadata } from "next";
import { ResumeUploader } from "@/components/resume-uploader";
export const metadata: Metadata = { title: "Resumes" };
export default function ResumesPage() {
  return (
    <>
      <div className="eyebrow">Documents</div>
      <h1
        className="title"
        style={{ fontSize: "clamp(2rem,4vw,3rem)", margin: "10px 0" }}
      >
        Your resumes
      </h1>
      <p
        className="muted"
        style={{ maxWidth: 660, lineHeight: 1.7, marginBottom: 30 }}
      >
        Keep the source documents that represent your real experience. Roliq
        verifies every upload before it can enter a later processing workflow.
      </p>
      <ResumeUploader />
    </>
  );
}
