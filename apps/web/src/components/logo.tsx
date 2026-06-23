import Link from "next/link";

export function Logo() {
  return (
    <Link
      href="/"
      aria-label="Roliq home"
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 10,
        fontWeight: 800,
        fontSize: "1.15rem",
        letterSpacing: "-.03em",
      }}
    >
      <span
        aria-hidden
        style={{
          display: "grid",
          placeItems: "center",
          width: 32,
          height: 32,
          borderRadius: 11,
          background: "#3157d5",
          color: "white",
          fontFamily: "Georgia",
          fontSize: 20,
        }}
      >
        R
      </span>
      Roliq
    </Link>
  );
}
