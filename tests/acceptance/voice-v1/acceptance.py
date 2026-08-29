#!/usr/bin/env python3

import base64
import hashlib
import hmac
import json
import os
import ssl
import subprocess
import sys
import time
import urllib.error
import urllib.request
import uuid


API_BASE = os.environ.get("VOICE_V1_API_BASE", "http://127.0.0.1:8080")
TOKEN = os.environ.get(
    "VOICE_V1_TOKEN",
    "lm_org_v1smoke0_v1smoke0abcdefghijklmnopqrstuvwx",
)
ESL_PASSWORD = os.environ.get("FREESWITCH_ESL_PASSWORD", "voice-v1-esl-secret")
DID = os.environ.get("VOICE_V1_DID", "+15551234567")
CALLER = os.environ.get("VOICE_V1_CALLER", "+15557654321")
PROVIDER_ID = "00000000-0000-0000-0000-000000002001"
COMPOSE = [
    "docker",
    "compose",
    "-f",
    "deploy/compose.yaml",
    "-f",
    "tests/acceptance/voice-v1/compose.yaml",
]


class AcceptanceError(RuntimeError):
    pass


class Results:
    def __init__(self):
        self.items = {}

    def pass_(self, number, name, detail=""):
        self.items[number] = ("PASS", name, detail)
        print(f"PASS {number:02d} {name}{': ' + detail if detail else ''}")

    def fail(self, number, name, error):
        self.items[number] = ("FAIL", name, str(error))
        print(f"FAIL {number:02d} {name}: {error}")

    def summary(self):
        print("\nVoice v1 acceptance matrix")
        print("=" * 72)
        failed = 0
        for number in range(1, 17):
            status, name, detail = self.items.get(
                number,
                ("FAIL", "unexecuted acceptance item", "dependency prevented execution"),
            )
            failed += status != "PASS"
            suffix = f" - {detail}" if detail else ""
            print(f"{number:02d}. {status:4} {name}{suffix}")
        return failed


RESULTS = Results()
STATE = {}


def run(command, *, check=True, input_text=None):
    completed = subprocess.run(
        command,
        input=input_text,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        check=False,
    )
    if check and completed.returncode != 0:
        raise AcceptanceError(
            f"command failed ({completed.returncode}): {' '.join(command)}\n{completed.stdout}"
        )
    return completed.stdout.strip()


def compose(*args, check=True):
    return run(COMPOSE + list(args), check=check)


def fs_cli(service, command):
    return compose(
        "exec",
        "-T",
        service,
        "fs_cli",
        "-H",
        "127.0.0.1",
        "-P",
        "8021",
        "-p",
        ESL_PASSWORD,
        "-x",
        command,
    )


def psql(sql):
    return compose(
        "exec",
        "-T",
        "postgres",
        "psql",
        "-v",
        "ON_ERROR_STOP=1",
        "-U",
        "leamout",
        "-d",
        "leamout",
        "-Atc",
        sql,
    )


def request(method, path, payload=None, *, auth=True, expected=None):
    body = None
    headers = {"Accept": "application/json"}
    if payload is not None:
        body = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"
    if auth:
        headers["Authorization"] = f"Bearer {TOKEN}"

    req = urllib.request.Request(
        API_BASE + path,
        data=body,
        method=method,
        headers=headers,
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            status = response.status
            raw = response.read()
    except urllib.error.HTTPError as error:
        status = error.code
        raw = error.read()
    except OSError as error:
        raise AcceptanceError(f"{method} {path}: {error}") from error

    if expected is not None and status not in expected:
        text = raw.decode("utf-8", errors="replace")
        raise AcceptanceError(f"{method} {path}: HTTP {status}, body={text}")

    if not raw:
        return status, None
    try:
        return status, json.loads(raw)
    except json.JSONDecodeError:
        return status, raw.decode("utf-8", errors="replace")


def wait_for(description, probe, timeout=15.0, interval=0.25):
    deadline = time.monotonic() + timeout
    last_error = None
    while time.monotonic() < deadline:
        try:
            value = probe()
            if value:
                return value
        except Exception as error:  # acceptance polling preserves the last failure
            last_error = error
        time.sleep(interval)
    suffix = f": {last_error}" if last_error else ""
    raise AcceptanceError(f"timed out waiting for {description}{suffix}")


def health_status(path):
    status, _ = request("GET", path, auth=False)
    return status


def wait_api_ready():
    wait_for("API liveness", lambda: health_status("/healthz") == 204, timeout=45)
    wait_for("API readiness", lambda: health_status("/readyz") == 204, timeout=45)


def get_call(call_id):
    _, call = request("GET", f"/v1/calls/{call_id}", expected={200})
    return call


def wait_call(call_id, predicate, description, timeout=15):
    return wait_for(
        description,
        lambda: (call := get_call(call_id)) if predicate(call) else False,
        timeout=timeout,
    )


def list_calls():
    _, body = request("GET", "/v1/calls/?limit=100", expected={200})
    return body["calls"]


def list_recordings():
    _, body = request("GET", "/v1/recordings/?limit=100", expected={200})
    return body["recordings"]


def webhook_events():
    context = ssl._create_unverified_context()
    req = urllib.request.Request("https://127.0.0.1:18443/events")
    with urllib.request.urlopen(req, timeout=5, context=context) as response:
        return json.loads(response.read())["events"]


def verify_webhook_signature(event, secret):
    headers = event["headers"]
    timestamp = headers["X-Leamout-Timestamp"]
    signature = headers["X-Leamout-Signature"]
    body = event["body"]
    secret_bytes = base64.urlsafe_b64decode(secret + "=" * (-len(secret) % 4))
    digest = hmac.new(
        secret_bytes,
        f"{timestamp}.{body}".encode("utf-8"),
        hashlib.sha256,
    ).hexdigest()
    expected = "v1=" + digest
    if not hmac.compare_digest(signature, expected):
        raise AcceptanceError("webhook signature did not verify")


def check(number, name, function):
    try:
        detail = function() or ""
        RESULTS.pass_(number, name, detail)
        return True
    except Exception as error:
        RESULTS.fail(number, name, error)
        return False


def acceptance_deploy():
    wait_api_ready()
    running = set(compose("ps", "--status", "running", "--services").splitlines())
    required = {
        "postgres",
        "redis",
        "nats",
        "api",
        "worker",
        "opensips",
        "rtpengine",
        "freeswitch",
        "voice-v1-carrier",
        "voice-v1-webhook",
    }
    missing = sorted(required - running)
    if missing:
        raise AcceptanceError("services not running: " + ", ".join(missing))
    return "full control/media stack is running"


def acceptance_provider():
    _, providers = request("GET", "/v1/carrier-providers/", expected={200})
    provider_items = providers.get("carrier_providers", providers.get("providers", []))
    if not any(item["id"] == PROVIDER_ID for item in provider_items):
        raise AcceptanceError("synthetic carrier provider fixture is not visible through API")

    _, connection = request(
        "POST",
        "/v1/carrier-connections/",
        {
            "provider_id": PROVIDER_ID,
            "name": "voice-v1-carrier",
            "inbound_enabled": True,
            "codecs": ["PCMU", "PCMA"],
        },
        expected={201},
    )
    STATE["connection_id"] = connection["id"]

    request(
        "POST",
        f"/v1/carrier-connections/{connection['id']}/source-ips",
        {"cidr": "172.30.0.50/32"},
        expected={201},
    )

    _, trunk = request(
        "POST",
        "/v1/trunks/",
        {
            "carrier_connection_id": connection["id"],
            "name": "voice-v1-trunk",
            "direction": "bidirectional",
        },
        expected={201},
    )
    STATE["trunk_id"] = trunk["id"]

    _, endpoint = request(
        "POST",
        f"/v1/trunks/{trunk['id']}/endpoints",
        {
            "host": "voice-v1-carrier",
            "port": 5060,
            "transport": "udp",
            "direction": "bidirectional",
        },
        expected={201},
    )
    if endpoint["host"] != "voice-v1-carrier":
        raise AcceptanceError("trunk endpoint was not persisted")
    return f"carrier connection {connection['id']} routes to synthetic SIP peer"


def acceptance_voice_application():
    _, number = request(
        "POST",
        "/v1/numbers/",
        {"number": DID, "country_code": "US", "voice_enabled": True},
        expected={201},
    )
    STATE["number_id"] = number["id"]

    connection_id = STATE["connection_id"]
    updated = psql(
        "UPDATE phone_numbers SET carrier_connection_id = "
        f"'{connection_id}'::uuid WHERE id = '{number['id']}'::uuid RETURNING id;"
    )
    if number["id"] not in updated:
        raise AcceptanceError("failed to assign acceptance DID to carrier connection")

    _, application = request(
        "POST",
        "/v1/voice-applications/",
        {"name": "voice-v1-acceptance", "caller_id": CALLER},
        expected={201},
    )
    STATE["application_id"] = application["id"]

    _, binding = request(
        "POST",
        f"/v1/voice-applications/{application['id']}/bindings",
        {"phone_number_id": number["id"]},
        expected={201},
    )
    if binding.get("phone_number_id") != number["id"]:
        raise AcceptanceError("voice application DID binding was not persisted")
    return f"voice application {application['id']} bound to {DID}"


def configure_webhook():
    _, created = request(
        "POST",
        "/v1/webhooks/",
        {
            "url": "https://voice-v1-webhook:8443/events",
            "subscribed_events": [
                "call.initiated",
                "call.answered",
                "call.held",
                "call.resumed",
                "call.completed",
                "call.ringing",
                "recording.started",
                "recording.completed",
            ],
        },
        expected={201},
    )
    STATE["webhook_id"] = created["webhook"]["id"]
    STATE["webhook_secret"] = created["signing_secret"]

    _, test_result = request(
        "POST",
        f"/v1/webhooks/{STATE['webhook_id']}/test",
        expected={200},
    )
    if test_result["response_status"] != 204:
        raise AcceptanceError(f"webhook test returned {test_result['response_status']}")


def acceptance_outbound():
    _, call = request(
        "POST",
        "/v1/calls/",
        {
            "application_id": STATE["application_id"],
            "trunk_id": STATE["trunk_id"],
            "from": CALLER,
            "to": DID,
        },
        expected={201},
    )
    if call["direction"] != "outbound" or not call.get("sip_call_id"):
        raise AcceptanceError("outbound call did not expose expected routing identity")
    STATE["call_id"] = call["id"]
    STATE["sip_call_id"] = call["sip_call_id"]
    return f"outbound call {call['id']} reached synthetic carrier"


def acceptance_answer():
    call_id = STATE["call_id"]
    _, answered = request("POST", f"/v1/calls/{call_id}/answer", expected={200})
    if answered["state"] not in {"answered", "active"}:
        raise AcceptanceError(f"answer state is {answered['state']}")
    STATE["answered"] = True
    return "answer persisted connected call state"


def acceptance_hold_resume():
    call_id = STATE["call_id"]
    request("POST", f"/v1/calls/{call_id}/hold", expected={200})
    wait_call(call_id, lambda call: call["media_state"] == "held", "held media state")
    request("POST", f"/v1/calls/{call_id}/unhold", expected={200})
    wait_call(call_id, lambda call: call["media_state"] == "active", "active media state")
    return "media_state transitions active -> held -> active"


def acceptance_play():
    call_id = STATE["call_id"]
    request(
        "POST",
        f"/v1/calls/{call_id}/play",
        {"path": "tone_stream://%(500,0,440)"},
        expected={200},
    )
    time.sleep(0.2)
    request("POST", f"/v1/calls/{call_id}/stop", expected={200})
    if "+OK" not in fs_cli("freeswitch", f"uuid_exists {STATE['sip_call_id']}"):
        raise AcceptanceError("media channel disappeared during playback")
    return "FreeSWITCH accepted playback and stop on live channel"


def acceptance_record():
    call_id = STATE["call_id"]
    path = "/var/lib/freeswitch/recordings/voice-v1-acceptance.wav"
    request(
        "POST",
        f"/v1/calls/{call_id}/record",
        {"action": "start", "path": path},
        expected={200},
    )
    recording = wait_for(
        "recording.started persistence",
        lambda: next(
            (
                item
                for item in list_recordings()
                if item["call_id"] == call_id and item["status"] == "recording"
            ),
            False,
        ),
        timeout=15,
    )
    request(
        "POST",
        f"/v1/calls/{call_id}/record",
        {"action": "stop", "path": path},
        expected={200},
    )
    wait_for(
        "recording.completed persistence",
        lambda: next(
            (
                item
                for item in list_recordings()
                if item["id"] == recording["id"] and item["status"] == "completed"
            ),
            False,
        ),
        timeout=15,
    )
    return f"recording {recording['id']} completed through lifecycle events"


def acceptance_transfer():
    call_id = STATE["call_id"]
    request(
        "POST",
        f"/v1/calls/{call_id}/transfer",
        {"destination": "9196", "dialplan": "XML", "context": "leamout"},
        expected={200},
    )
    time.sleep(0.25)
    output = fs_cli("freeswitch", f"uuid_exists {STATE['sip_call_id']}")
    if "+OK true" not in output.lower() and "true" not in output.lower():
        raise AcceptanceError("transferred channel is no longer live")
    return "live call transferred into local 9196 dialplan"


def acceptance_hangup():
    call_id = STATE["call_id"]
    _, ended = request("POST", f"/v1/calls/{call_id}/hangup", expected={200})
    if ended["state"] not in {"completed", "cancelled"}:
        raise AcceptanceError(f"hangup state is {ended['state']}")
    STATE["terminal_state"] = ended["state"]
    return f"hangup persisted terminal state {ended['state']}"


def acceptance_inbound():
    existing_ids = {item["id"] for item in list_calls()}
    output = fs_cli(
        "voice-v1-carrier",
        "originate "
        f"{{origination_caller_id_number={CALLER}}}"
        f"sofia/internal/{DID}@opensips:5060 &park()",
    )
    if "+OK" not in output:
        raise AcceptanceError(f"carrier originate did not answer: {output}")
    carrier_uuid = output.split("+OK", 1)[1].strip().split()[0]

    inbound = wait_for(
        "inbound call persistence",
        lambda: next(
            (
                item
                for item in list_calls()
                if item["id"] not in existing_ids
                and item["direction"] == "inbound"
                and item["to"] == DID
            ),
            False,
        ),
        timeout=15,
    )
    STATE["inbound_call_id"] = inbound["id"]
    fs_cli("voice-v1-carrier", f"uuid_kill {carrier_uuid}")
    wait_call(
        inbound["id"],
        lambda call: call["state"] in {"completed", "failed", "cancelled"},
        "inbound terminal state",
        timeout=15,
    )
    return f"carrier ingress created inbound call {inbound['id']}"


def acceptance_conference():
    name = "voice-v1-" + uuid.uuid4().hex[:8]
    _, conference = request(
        "POST",
        "/v1/conferences/",
        {"application_id": STATE["application_id"], "name": name},
        expected={201},
    )
    if conference["state"] != "active":
        raise AcceptanceError("conference API did not create active conference")

    media = fs_cli("freeswitch", "conference list")
    if name not in media:
        raise AcceptanceError(
            "conference exists only in control-plane state; no matching FreeSWITCH conference exists"
        )

    request("POST", f"/v1/conferences/{conference['id']}/lock", expected={200})
    media = fs_cli("freeswitch", f"conference {name} list")
    if "locked" not in media.lower():
        raise AcceptanceError("conference lock is not observable in FreeSWITCH")
    request("POST", f"/v1/conferences/{conference['id']}/unlock", expected={200})
    request("DELETE", f"/v1/conferences/{conference['id']}", expected={200})
    return "conference lifecycle and media controls are backed by FreeSWITCH"


def acceptance_events():
    required = {"call.initiated", "call.answered", "call.held", "call.resumed", "call.completed"}

    def observed():
        types = {event["envelope"].get("type") for event in webhook_events()}
        return types if required.issubset(types) else False

    types = wait_for("normalized call events", observed, timeout=20)
    return "observed " + ", ".join(sorted(required & types))


def acceptance_query_state():
    call = get_call(STATE["call_id"])
    if call["state"] != STATE["terminal_state"]:
        raise AcceptanceError("queried call state does not match last control mutation")
    if call["organization_id"] != "00000000-0000-0000-0000-000000001001":
        raise AcceptanceError("queried call escaped acceptance organization")
    return f"GET /calls/{{id}} returned durable {call['state']} state"


def acceptance_webhook():
    events = wait_for("webhook deliveries", lambda: webhook_events() or False, timeout=20)
    call_events = [event for event in events if event["envelope"].get("type", "").startswith("call.")]
    if not call_events:
        raise AcceptanceError("no call webhook deliveries were received")
    for event in call_events:
        verify_webhook_signature(event, STATE["webhook_secret"])
    _, deliveries = request(
        "GET",
        f"/v1/webhooks/{STATE['webhook_id']}/deliveries?limit=100",
        expected={200},
    )
    if not any(item["status"] == "delivered" for item in deliveries["deliveries"]):
        raise AcceptanceError("webhook API has no delivered attempt")
    return f"{len(call_events)} signed call webhook deliveries verified"


def acceptance_health():
    if health_status("/healthz") != 204 or health_status("/readyz") != 204:
        raise AcceptanceError("HTTP liveness/readiness is not healthy")
    status = fs_cli("freeswitch", "status")
    if "UP" not in status.upper():
        raise AcceptanceError("FreeSWITCH status is not UP")
    channels = fs_cli("freeswitch", "show channels count")
    if "total" not in channels.lower() and "+OK" not in channels:
        raise AcceptanceError("FreeSWITCH channel health query failed")
    return "HTTP readiness and FreeSWITCH media status are healthy"


def acceptance_restart():
    call_before = get_call(STATE["call_id"])

    compose("restart", "worker")
    wait_for(
        "worker restart",
        lambda: "worker" in compose("ps", "--status", "running", "--services").splitlines(),
        timeout=30,
    )
    compose("restart", "api")
    wait_api_ready()
    after_process_restart = get_call(STATE["call_id"])
    if after_process_restart["state"] != call_before["state"]:
        raise AcceptanceError("API/worker restart changed terminal call state")

    compose("stop", "freeswitch")
    wait_for("readiness failure after FreeSWITCH stop", lambda: health_status("/readyz") == 503, timeout=15)
    compose("start", "freeswitch")
    wait_for("readiness recovery after FreeSWITCH start", lambda: health_status("/readyz") == 204, timeout=45)

    after_media_restart = get_call(STATE["call_id"])
    if after_media_restart["state"] != call_before["state"]:
        raise AcceptanceError("FreeSWITCH restart changed completed call state")
    return "API, worker, and FreeSWITCH restarted with readiness recovery and durable state"


def main():
    if not check(1, "Deploy Leamout", acceptance_deploy):
        RESULTS.summary()
        return 1

    if not check(2, "Configure a SIP endpoint/provider", acceptance_provider):
        RESULTS.summary()
        return 1

    if not check(3, "Create a voice application", acceptance_voice_application):
        RESULTS.summary()
        return 1

    try:
        configure_webhook()
    except Exception as error:
        RESULTS.fail(14, "Receive webhooks", f"webhook setup failed: {error}")

    check(4, "Receive an inbound call", acceptance_inbound)

    outbound_ok = check(5, "Originate an outbound call", acceptance_outbound)
    if outbound_ok:
        answer_ok = check(6, "Answer/hang up", acceptance_answer)
        if answer_ok:
            check(8, "Hold/resume", acceptance_hold_resume)
            check(9, "Play audio", acceptance_play)
            check(10, "Record", acceptance_record)
            check(7, "Transfer", acceptance_transfer)
            try:
                detail = acceptance_hangup()
                if RESULTS.items.get(6, (None,))[0] == "PASS":
                    RESULTS.pass_(6, "Answer/hang up", detail)
            except Exception as error:
                RESULTS.fail(6, "Answer/hang up", f"answer passed but hangup failed: {error}")

    check(11, "Create/manage conferences", acceptance_conference)
    if 14 not in RESULTS.items or RESULTS.items[14][0] != "FAIL":
        check(14, "Receive webhooks", acceptance_webhook)
    check(12, "Receive normalized call events", acceptance_events)
    check(13, "Query call state", acceptance_query_state)
    check(15, "Inspect call/media health", acceptance_health)
    check(16, "Restart components without corrupting state", acceptance_restart)

    failed = RESULTS.summary()
    if failed:
        print(f"\nVoice v1 acceptance FAILED: {failed} capability check(s) did not pass.")
        return 1
    print("\nVoice v1 acceptance PASSED: all 16 capabilities are complete.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
