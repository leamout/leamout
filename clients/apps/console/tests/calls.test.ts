import assert from "node:assert/strict";
import { test } from "node:test";
import { api } from "../lib/api";
import {
  callDurationSeconds,
  formatCallDate,
  formatDuration,
  type Call,
  type CallRecording,
} from "../components/calls/types";

const baseCall: Call = {
  id: "call-a",
  organization_id: "org-a",
  voice_agent_id: "agent-a",
  trunk_id: "trunk-a",
  direction: "inbound",
  state: "completed",
  media_state: "active",
  from_uri: "sip:caller@example.com",
  to_uri: "sip:callee@example.com",
  sip_call_id: "sip-a",
  started_at: "2026-10-23T11:59:50Z",
  answered_at: "2026-10-23T12:00:00Z",
  ended_at: "2026-10-23T12:01:05Z",
  created_at: "2026-10-23T11:59:50Z",
  updated_at: "2026-10-23T12:01:05Z",
};

test("Call list, detail, and recording filters use organization scoped GETs", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ path: string; options: RequestInit }> = [];

  globalThis.fetch = async (path, options) => {
    const target = String(path);
    requests.push({ path: target, options: options ?? {} });
    if (target.includes("/recordings/")) {
      return Response.json({
        success: true,
        data: { recordings: [] as CallRecording[] },
      });
    }
    return Response.json({
      success: true,
      data: target.includes("/calls/?")
        ? { calls: [baseCall] }
        : baseCall,
    });
  };

  try {
    const params = new URLSearchParams({
      offset: "50",
      limit: "50",
      state: "failed",
      direction: "outbound",
      voice_agent_id: "agent-a",
    });

    const list = await api<{ calls: Call[] }>(
      "/calls/?" + params.toString(),
      { organizationId: "org-a" },
    );
    const detail = await api<Call>("/calls/call-a", {
      organizationId: "org-a",
    });
    const recordings = await api<{ recordings: CallRecording[] }>(
      "/recordings/?call_id=call-a&limit=50&offset=0",
      { organizationId: "org-a" },
    );

    assert.equal(list.calls[0].id, "call-a");
    assert.equal(detail.sip_call_id, "sip-a");
    assert.deepEqual(recordings.recordings, []);
    assert.equal(requests.length, 3);

    const listPath = new URL(requests[0].path, "https://example.test");
    assert.equal(listPath.pathname, "/api/v1/calls/");
    assert.equal(listPath.searchParams.get("offset"), "50");
    assert.equal(listPath.searchParams.get("limit"), "50");
    assert.equal(listPath.searchParams.get("state"), "failed");
    assert.equal(listPath.searchParams.get("direction"), "outbound");
    assert.equal(listPath.searchParams.get("voice_agent_id"), "agent-a");
    assert.equal(requests[1].path, "/api/v1/calls/call-a");
    assert.equal(
      requests[2].path,
      "/api/v1/recordings/?call_id=call-a&limit=50&offset=0",
    );

    for (const request of requests) {
      const headers = new Headers(request.options.headers);
      assert.equal(request.options.method, "GET");
      assert.equal(request.options.credentials, "same-origin");
      assert.equal(request.options.cache, "no-store");
      assert.equal(headers.get("X-Organization-ID"), "org-a");
      assert.equal(headers.get("Authorization"), null);
    }
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Call duration uses only persisted answered and ended timestamps", () => {
  assert.equal(callDurationSeconds(baseCall), 65);
  assert.equal(formatDuration(65), "1m 5s");
  assert.equal(formatDuration(3661), "1h 1m 1s");
  assert.equal(formatDuration(null), "—");
  assert.equal(
    callDurationSeconds({ ...baseCall, answered_at: null }),
    null,
  );
  assert.equal(
    callDurationSeconds({ ...baseCall, ended_at: null, state: "active" }),
    null,
  );
  assert.equal(
    callDurationSeconds({
      ...baseCall,
      answered_at: "2026-10-23T12:10:00Z",
    }),
    0,
  );
  assert.equal(formatCallDate("invalid-date"), "—");
  assert.equal(formatCallDate(null), "—");
});
