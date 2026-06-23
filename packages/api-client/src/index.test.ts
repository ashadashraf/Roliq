import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, RoliqClient } from "./index";

afterEach(() => vi.unstubAllGlobals());
describe("RoliqClient", () => {
  it("adds bearer authorization and decodes problem responses", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          code: "validation_failed",
          detail: "Review fields",
          requestId: "req-1",
        }),
        {
          status: 422,
          headers: { "content-type": "application/problem+json" },
        },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);
    const client = new RoliqClient(
      "https://api.roliq.test",
      async () => "token",
    );
    await expect(client.getProfile()).rejects.toMatchObject({
      status: 422,
      code: "validation_failed",
      requestId: "req-1",
    });
    expect(fetchMock.mock.calls[0][1].headers.get("Authorization")).toBe(
      "Bearer token",
    );
  });
});
