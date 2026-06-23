"use client";

import { useAuth } from "@clerk/nextjs";
import { RoliqClient } from "@roliq/api-client";
import { useMemo } from "react";

export function useRoliqClient() {
  const { getToken } = useAuth();
  return useMemo(
    () =>
      new RoliqClient(
        process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080",
        () => getToken({ template: "roliq-api" }),
      ),
    [getToken],
  );
}
