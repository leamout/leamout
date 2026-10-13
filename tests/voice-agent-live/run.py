#!/usr/bin/env python3
"""Explicitly opted-in, manually witnessed live BYOC/BYOA golden-call validation.

Unlike voice-agent-v1, this runs against a real carrier and a real AI provider.
It never provisions or changes customers, trunks, agents, or credentials.
"""
import argparse
import datetime as dt
import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid

TERMINAL = {"completed", "failed", "cancelled"}


class AcceptanceError(RuntimeError):
    pass


def require_env(name):
    value = os.getenv(name, "").strip()
    if not value:
        raise AcceptanceError(f"missing required {name}")
    return value


def require_uuid(name):
    try:
        value = uuid.UUID(require_env(name))
    except ValueError as error:
        raise AcceptanceError(f"{name} must be a valid UUID") from error
    if value.int == 0:
        raise AcceptanceError(f"{name} may not be the nil UUID")
    return str(value)


def validate_origin(origin):
    parsed = urllib.parse.urlsplit(origin.rstrip("/"))
    local = parsed.hostname in {"localhost", "127.0.0.1", "::1"}
    if (not parsed.netloc or parsed.username or parsed.password or parsed.query
            or parsed.fragment or parsed.path not in {"", "/"}
            or not (parsed.scheme == "https" or (parsed.scheme == "http" and local))):
        raise AcceptanceError("API origin must be HTTPS (HTTP permitted only for loopback)")
    return origin.rstrip("/")


def validate_media(session):
    values = [session.get(name) for name in
              ("media_started_at", "media_ended_at", "ended_at")]
    if not all(values):
        raise AcceptanceError("durable media start/end timestamps are missing")
    try:
        start, stop, end = (
            dt.datetime.fromisoformat(value.replace("Z", "+00:00"))
            for value in values
        )
    except ValueError:
        raise AcceptanceError("invalid media timestamp") from None
    if not start < stop <= end:
        raise AcceptanceError("media interval must be positive and bounded by session end")


def wait_for(description, probe, timeout, interval=2):
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        value = probe()
        if value:
            return value
        time.sleep(interval)
    raise AcceptanceError(f"timed out waiting for {description}")


class Runner:
    def __init__(self, args):
        self.args = args
        if os.getenv("LEAMOUT_LIVE_CONFIRM") != "YES":
            raise AcceptanceError("set LEAMOUT_LIVE_CONFIRM=YES to acknowledge live carrier/AI costs")
        if not sys.stdin.isatty():
            raise AcceptanceError("live validation requires an interactive operator for audio confirmation")
        self.api_base = validate_origin(require_env("LEAMOUT_LIVE_API_BASE"))
        self.token = require_env("LEAMOUT_LIVE_TOKEN")
        self.org = require_uuid("LEAMOUT_LIVE_ORGANIZATION_ID")
        self.agent = require_uuid("LEAMOUT_LIVE_AGENT_ID")
        self.trunk = require_uuid("LEAMOUT_LIVE_TRUNK_ID")
        self.inbound_to = require_env("LEAMOUT_LIVE_INBOUND_TO") if args.direction != "outbound" else ""
        self.outbound_from = require_env("LEAMOUT_LIVE_OUTBOUND_FROM") if args.direction != "inbound" else ""
        self.outbound_to = require_env("LEAMOUT_LIVE_OUTBOUND_TO") if args.direction != "inbound" else ""
        for name in ("PGHOST", "PGDATABASE", "PGUSER"):
            require_env(name)
        if os.environ["PGHOST"] not in {"localhost", "127.0.0.1", "::1"}:
            if os.getenv("PGSSLMODE") not in {"require", "verify-ca", "verify-full"}:
                raise AcceptanceError("remote PostgreSQL requires PGSSLMODE=require or stronger")
        self.outbound_call = None

    def api(self, method, path, payload=None, key=None):
        headers = {"Accept": "application/json", "Authorization": "Bearer " + self.token}
        data = None
        if payload is not None:
            data = json.dumps(payload).encode()
            headers["Content-Type"] = "application/json"
        if key:
            headers["Idempotency-Key"] = key
        request = urllib.request.Request(
            self.api_base + path, data=data, method=method, headers=headers
        )
        try:
            with urllib.request.urlopen(request, timeout=12) as response:
                status, raw = response.status, response.read(1 << 20)
        except urllib.error.HTTPError as error:
            # Never leak provider, phone number, token or transcript response bodies.
            raise AcceptanceError(f"{method} {path.split('?')[0]}: HTTP {error.code}") from None
        except urllib.error.URLError:
            raise AcceptanceError(f"{method} API endpoint is unavailable") from None
        wanted = 201 if method == "POST" and path == "/v1/calls/" else 200
        if status != wanted:
            raise AcceptanceError(f"{method} API response was HTTP {status}, expected {wanted}")
        try:
            envelope = json.loads(raw)
        except (ValueError, UnicodeDecodeError):
            raise AcceptanceError("API returned invalid JSON") from None
        if not isinstance(envelope, dict) or envelope.get("success") is not True:
            raise AcceptanceError("API returned an unsuccessful response")
        result = envelope.get("data")
        if not isinstance(result, dict):
            raise AcceptanceError("API response data must be an object")
        return result

    def sql_json(self, query):
        # libpq reads PGHOST/PGUSER/PGDATABASE/PGPASSWORD without secrets in argv.
        try:
            result = subprocess.run(
                ["psql", "-X", "-q", "-A", "-t", "--no-password",
                 "-v", "ON_ERROR_STOP=1", "-c", query],
                capture_output=True, text=True, timeout=15, check=False
            )
        except (OSError, subprocess.TimeoutExpired):
            raise AcceptanceError("psql is missing or did not respond") from None
        if result.returncode:
            # Do not echo potentially sensitive PostgreSQL stderr.
            raise AcceptanceError("PostgreSQL verification query failed")
        try:
            return json.loads(result.stdout.strip())
        except ValueError:
            raise AcceptanceError("PostgreSQL verification query returned invalid JSON") from None

    def list_db(self, table, columns, where):
        return self.sql_json(
            "SELECT COALESCE(json_agg(row_to_json(t)), '[]'::json)::text FROM "
            f"(SELECT {columns} FROM {table} WHERE {where}) t"
        )

    def preflight(self):
        agent = self.api("GET", f"/v1/voice-agents/{self.agent}")
        if (agent.get("id") != self.agent or agent.get("organization_id") != self.org
                or agent.get("status") != "active"):
            raise AcceptanceError("Voice Agent must be active in the selected organization")
        readiness = self.api("GET", f"/v1/voice-agents/{self.agent}/readiness")
        if not readiness.get("ready") or not readiness.get("bindings") or any(
                item.get("connection_state") != "ready" for item in readiness["bindings"]):
            raise AcceptanceError("Voice Agent BYOA bindings are not all ready")
        self.api("GET", "/v1/calls/?limit=1")
        self.sql_json("SELECT to_jsonb(COUNT(*))::text FROM organizations")
        print("PASS preflight: API, verified real BYOA bindings, PostgreSQL")

    def check_call(self, call, direction):
        if (call.get("organization_id") != self.org
                or call.get("voice_agent_id") != self.agent
                or call.get("trunk_id") != self.trunk
                or call.get("direction") != direction):
            raise AcceptanceError("unexpected organization/agent/trunk on real call")
        try:
            return str(uuid.UUID(call["id"]))
        except (ValueError, TypeError, KeyError):
            raise AcceptanceError("live call ID is missing or invalid") from None

    def inbound(self):
        since = dt.datetime.now(dt.timezone.utc).isoformat(timespec="seconds")
        print("Place ONE inbound call through your real SIP carrier/PBX NOW.")
        print("Talk to the agent, listen for its reply, then hang up.")
        path = "/v1/calls/?" + urllib.parse.urlencode({
            "direction": "inbound", "voice_agent_id": self.agent,
            "trunk_id": self.trunk, "created_from": since, "limit": 100,
        })
        def probe():
            for call in self.api("GET", path).get("calls", []):
                if (call.get("organization_id") == self.org
                        and call.get("voice_agent_id") == self.agent
                        and call.get("trunk_id") == self.trunk
                        and self.inbound_to in call.get("to_uri", "")):
                    return self.check_call(call, "inbound")
            return None
        return wait_for("new real inbound call", probe, self.args.timeout)

    def outbound(self):
        print("Starting exactly ONE paid outbound call. Answer and speak with the agent.")
        call = self.api("POST", "/v1/calls/", {
            "voice_agent_id": self.agent, "trunk_id": self.trunk,
            "from_uri": self.outbound_from, "to_uri": self.outbound_to,
        }, key="leamout-live-" + str(uuid.uuid4()))
        self.outbound_call = self.check_call(call, "outbound")
        print("After hearing the agent reply, hang up.")
        return self.outbound_call

    def completed_call(self, call_id):
        call = wait_for("real call to end", lambda: self.probe_terminal(call_id), self.args.timeout)
        if call.get("state") != "completed" or not call.get("answered_at") or not call.get("ended_at"):
            raise AcceptanceError("call did not end in successful completed/answered state")
        return call

    def probe_terminal(self, call_id):
        call = self.api("GET", f"/v1/calls/{call_id}")
        return call if call.get("state") in TERMINAL else None

    def durable(self, call_id):
        where = f"organization_id = '{self.org}'::uuid AND call_id = '{call_id}'::uuid"
        def session_probe():
            rows = self.list_db("voice_agent_sessions",
                                "id::text, state, media_started_at, media_ended_at, "
                                "ended_at, interruption_count", where)
            if len(rows) > 1:
                raise AcceptanceError("more than one durable Voice Agent session per call")
            return rows[0] if rows and rows[0]["state"] != "active" else None
        session = wait_for("terminal durable Voice Agent session", session_probe, 45)
        if session["state"] != "completed":
            raise AcceptanceError("durable Voice Agent session did not complete")
        validate_media(session)
        session_id = str(uuid.UUID(session["id"]))
        turn_where = f"organization_id = '{self.org}'::uuid AND session_id = '{session_id}'::uuid"
        def turns_probe():
            rows = self.list_db("voice_agent_turns", "sequence, role", turn_where + " ORDER BY sequence")
            return rows if {"user", "assistant"}.issubset({r["role"] for r in rows}) else None
        turns = wait_for("persisted human/assistant turns", turns_probe, 30)
        sequences = [row["sequence"] for row in turns]
        if not sequences or sequences[0] != 1 or sequences != sorted(set(sequences)):
            raise AcceptanceError("durable conversation turns are not in order")
        if self.args.require_tool and not any(
            row["state"] == "succeeded" for row in self.list_db(
                "voice_agent_tool_executions", "state", turn_where)
        ):
            raise AcceptanceError("successful durable tool execution missing")
        if self.args.require_barge_in and session["interruption_count"] < 1:
            raise AcceptanceError("barge-in was not recorded")
        if self.args.require_recording and not any(
            row["status"] == "completed" and (row["file_size_bytes"] or 0) > 0
            for row in self.list_db(
                "recordings", "status, file_size_bytes", where)
        ):
            raise AcceptanceError("completed nonempty recording missing")
        print("PASS session completed, nonzero media interval, ordered user/assistant turns")
        if self.args.require_tool:
            print("PASS durable tool call")
        if self.args.require_barge_in:
            print("PASS barge-in")
        if self.args.require_recording:
            print("PASS recording")

    def cleanup(self):
        if self.outbound_call is None:
            return
        try:
            call = self.api("GET", f"/v1/calls/{self.outbound_call}")
            if call.get("state") not in TERMINAL:
                self.api("POST", f"/v1/calls/{self.outbound_call}/hangup",
                         key="leamout-live-cleanup-" + str(uuid.uuid4()))
                print("Attempted cleanup of an active outbound call.")
        except AcceptanceError:
            print("WARNING: outbound call cleanup uncertain; inspect live call immediately.")
        self.outbound_call = None

    def run(self):
        self.preflight()
        directions = ["inbound", "outbound"] if self.args.direction == "both" else [self.args.direction]
        for direction in directions:
            try:
                call_id = self.inbound() if direction == "inbound" else self.outbound()
                self.completed_call(call_id)
                self.durable(call_id)
                print(f"{direction.upper()}: confirm that you personally heard BOTH sides of the live call.")
                if input("Type AUDIO-OK only if two-way audio was audible: ").strip() != "AUDIO-OK":
                    raise AcceptanceError("human operator did not confirm audible two-way audio")
                self.outbound_call = None
                print(f"PASS {direction}: real golden-call lifecycle with human audio witness")
            except (AcceptanceError, KeyboardInterrupt):
                self.cleanup()
                raise
        print("PASS live BYOC/BYOA golden-call validation")


def main(argv=None):
    parser = argparse.ArgumentParser(description="Manual live Voice Agent acceptance")
    parser.add_argument("--direction", choices=("inbound", "outbound", "both"), default="both")
    parser.add_argument("--timeout", type=int, default=240)
    parser.add_argument("--require-tool", action="store_true")
    parser.add_argument("--require-barge-in", action="store_true")
    parser.add_argument("--require-recording", action="store_true")
    args = parser.parse_args(argv)
    if not 30 <= args.timeout <= 900:
        parser.error("--timeout must be 30-900 seconds")
    try:
        Runner(args).run()
    except AcceptanceError as error:
        print(f"FAIL live Voice Agent golden call: {error}", file=sys.stderr)
        return 1
    except KeyboardInterrupt:
        print("Interrupted by operator", file=sys.stderr)
        return 130
    return 0


if __name__ == "__main__":
    sys.exit(main())
