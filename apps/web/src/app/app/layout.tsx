import { BootstrapGate } from "@/components/bootstrap-gate";
import { AppShell } from "@/components/app-shell";
export default function AuthenticatedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <BootstrapGate>
      <AppShell>{children}</AppShell>
    </BootstrapGate>
  );
}
