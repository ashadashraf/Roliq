import type { Metadata } from "next";
import { BootstrapGate } from "@/components/bootstrap-gate";
import { OnboardingFlow } from "@/components/onboarding-flow";
export const metadata: Metadata = { title: "Set up your workspace" };
export default function OnboardingPage() {
  return (
    <BootstrapGate>
      <OnboardingFlow />
    </BootstrapGate>
  );
}
