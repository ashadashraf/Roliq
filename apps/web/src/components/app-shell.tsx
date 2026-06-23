"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { FileText, LayoutDashboard, UserRound } from "lucide-react";
import { UserButton } from "@clerk/nextjs";
import { Logo } from "./logo";

const links = [
  { href: "/app", label: "Overview", icon: LayoutDashboard },
  { href: "/app/profile", label: "Career profile", icon: UserRound },
  { href: "/app/resumes", label: "Resumes", icon: FileText },
] as const;

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  return (
    <div
      className="app-layout"
      style={{
        minHeight: "100vh",
        display: "grid",
        gridTemplateColumns: "240px minmax(0,1fr)",
      }}
    >
      <aside
        className="app-aside"
        style={{
          borderRight: "1px solid var(--line)",
          background: "var(--surface)",
          padding: 24,
          display: "flex",
          flexDirection: "column",
          position: "sticky",
          top: 0,
          height: "100vh",
        }}
      >
        <Logo />
        <nav
          className="app-nav"
          aria-label="Workspace"
          style={{ display: "grid", gap: 7, marginTop: 42 }}
        >
          {links.map(({ href, label, icon: Icon }) => {
            const active =
              href === "/app" ? pathname === href : pathname.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 11,
                  padding: "11px 13px",
                  borderRadius: 11,
                  fontWeight: 650,
                  color: active ? "var(--brand)" : "#475467",
                  background: active ? "#eef2ff" : "transparent",
                }}
              >
                <Icon size={18} />
                {label}
              </Link>
            );
          })}
        </nav>
        <div
          style={{
            marginTop: "auto",
            display: "flex",
            alignItems: "center",
            gap: 10,
            paddingTop: 18,
            borderTop: "1px solid var(--line)",
          }}
        >
          <UserButton showName />
          <span className="muted" style={{ fontSize: 12 }}>
            Personal workspace
          </span>
        </div>
      </aside>
      <div style={{ minWidth: 0 }}>
        <header
          style={{
            height: 70,
            borderBottom: "1px solid var(--line)",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            padding: "0 34px",
            background: "rgba(247,245,239,.88)",
            backdropFilter: "blur(12px)",
            position: "sticky",
            top: 0,
            zIndex: 5,
          }}
        >
          <span style={{ fontWeight: 700 }}>Career workspace</span>
          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            <span className="pill">Private by default</span>
            <UserButton />
          </div>
        </header>
        <main
          style={{ padding: "38px clamp(20px,4vw,52px) 72px", maxWidth: 1220 }}
        >
          {children}
        </main>
      </div>
    </div>
  );
}
