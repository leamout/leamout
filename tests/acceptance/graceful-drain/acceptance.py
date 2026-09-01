#!/usr/bin/env python3
import json
import os
import re
import subprocess
import sys
import time
import urllib.error
import urllib.request

API = os.getenv("GRACEFUL_DRAIN_API_BASE", "http://127.0.0.1:8080")
TOKEN = os.getenv(
    "GRACEFUL_DRAIN_TOKEN",
    "lm_org_v1smoke0_v1smoke0abcdefghijklmnopqrstuvwx",
)
ESL_PASSWORD = os.getenv("FREESWITCH_ESL_PASSWORD", "graceful-drain-esl-secret")
DID = "+15551234567"
CALLER = "+15557654321"
COMPOSE = [
    "docker",
    "compose",
    "-f",
    "deploy/compose.yaml",
    "-f",
    "tests/acceptance/graceful-drain/compose.yaml",
]
STATE = {}


class Failure(RuntimeError):
    pass


def run(args, check=True):
    proc = subprocess.run(
        args,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )
    if check and proc.returncode:
        raise Failure(f"command failed: {' '.join(args)}\n{proc.stdout}")
    return proc.stdout.strip()


def compose(*args, check=True):
    return run(COMPOSE + list(args), check)


def fs(service, command):
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


def api(method, path, payload=None, expected=(200,)):
    data = json.dumps(payload).encode() if payload is not None else None
    headers = {
        "Accept": "application/json",
        "Authorization": f"Bearer {TOKEN}",
    }
    if data is not None:
        headers["Content-Type"] = "application/json"

    request = urllib.request.Request(
        API + path,
        data=data,
        headers=headers,
        method=method,
    )
    try:
        with urllib.request.urlopen(request, timeout=15) as response:
            status = response.status
            raw = response.read()
    except urllib.error.HTTPError as error:
        status = error.code
        raw = error.read()

    if status not in expected:
        raise Failure(
            f"{method} {path}: HTTP {status}, body={raw.decode(errors='replace')}"
        )
    if not raw:
        return None

    body = json.loads(raw)
    if isinstance(body, dict) and body.get("success") is True:
        return body.get("data")
    return body


def wait(description, probe, timeout=30):
    deadline = time.monotonic() + timeout
    last_error = None
    while time.monotonic() < deadline:
        try:
            value = probe()
            if value:
                return value
        except Exception as error:  # noqa: BLE001 - retain probe diagnostics
            last_error = error
        time.sleep(0.25)
    detail = f": {last_error}" if last_error else ""
    raise Failure(f"timed out waiting for {description}{detail}")


def numeric(output, description):
    value = output.strip()
    if not value.isdigit():
        raise Failure(f"{description} is not numeric: {output!r}")
    return int(value)


def channel_count(service):
    output = fs(service, "show channels count")
    match = re.search(r"(?m)^\s*(\d+)\s+total\.\s*$", output)
    if not match:
        raise Failure(f"unable to parse {service} channel count: {output!r}")
    return int(match.group(1))


def opensips_dialogs():
    return numeric(
        compose("exec", "-T", "opensips", "/usr/local/bin/leamout-opensips-drain", "dialogs"),
        "OpenSIPS dialog count",
    )


def freeswitch_channels():
    return numeric(
        compose("exec", "-T", "freeswitch", "/usr/local/bin/leamout-freeswitch-drain", "channels"),
        "FreeSWITCH channel count",
    )


def rtpengine_media_sockets():
    output = compose(
        "exec",
        "-T",
        "rtpengine",
        "sh",
        "-lc",
        "netstat -anu 2>/dev/null | awk 'NR > 2 { n=split($4,a,\":\"); p=a[n]+0; if (p >= 23000 && p <= 32768) count++ } END { print count+0 }'",
    )
    return numeric(output, "RTPengine media socket count")


def provision():
    providers = api("GET", "/v1/carrier-providers/")
    provider = next(
        (
            item
            for item in providers["carrier_providers"]
            if item["slug"] == "generic-sip" and item["status"] == "active"
        ),
        None,
    )
    if provider is None:
        raise Failure("generic-sip carrier provider is unavailable")

    connection = api(
        "POST",
        "/v1/carrier-connections/",
        {
            "provider_id": provider["id"],
            "name": "Graceful drain synthetic carrier",
            "inbound_enabled": True,
        },
        (201,),
    )
    STATE["connection"] = connection

    api(
        "POST",
        f"/v1/carrier-connections/{connection['id']}/source-ips",
        {"cidr": "172.30.0.50/32"},
        (201,),
    )

    number = api(
        "POST",
        "/v1/numbers/",
        {"number": DID, "country_code": "US", "voice_enabled": True},
        (201,),
    )
    STATE["number"] = number
    api(
        "PUT",
        f"/v1/numbers/{number['id']}/carrier-connection",
        {"carrier_connection_id": connection["id"]},
    )

    app = api(
        "POST",
        "/v1/voice-applications/",
        {"name": "Graceful drain ingress", "caller_id": CALLER},
        (201,),
    )
    STATE["app"] = app
    api(
        "POST",
        f"/v1/voice-applications/{app['id']}/bindings",
        {"phone_number_id": number["id"]},
        (201,),
    )


def establish_call():
    before = {
        call["id"]
        for call in api("GET", "/v1/calls/?limit=100")["calls"]
    }

    originate = fs(
        "graceful-drain-carrier",
        f"bgapi originate {{origination_caller_id_number={CALLER}}}sofia/internal/{DID}@opensips:5060 &park()",
    )
    if "+OK Job-UUID:" not in originate:
        raise Failure(f"carrier did not accept originate job: {originate}")

    def inbound_call():
        calls = api("GET", "/v1/calls/?limit=100")["calls"]
        return next(
            (
                call
                for call in calls
                if call["id"] not in before and call.get("direction") == "inbound"
            ),
            None,
        )

    call = wait("inbound call record", inbound_call, 30)
    STATE["call"] = call
    if not call.get("sip_call_id"):
        raise Failure("inbound call is missing SIP Call-ID attribution")

    answered = api("POST", f"/v1/calls/{call['id']}/answer")
    if answered.get("state") not in ("answered", "active"):
        raise Failure(f"answer returned unexpected state: {answered.get('state')!r}")

    def connected_call():
        current = api("GET", f"/v1/calls/{call['id']}")
        return current if current.get("state") in ("answered", "active") else None

    current = wait("answered call state", connected_call, 20)
    STATE["call"] = current

    wait("OpenSIPS dialog", lambda: opensips_dialogs() > 0, 15)
    wait("Leamout FreeSWITCH channel", lambda: freeswitch_channels() > 0, 15)
    wait(
        "synthetic carrier FreeSWITCH channel",
        lambda: channel_count("graceful-drain-carrier") > 0,
        15,
    )
    wait("RTPengine media sockets", lambda: rtpengine_media_sockets() >= 2, 15)

    print(
        "PASS established real SIP/media call "
        f"call={current['id']} sip_call_id={current['sip_call_id']} "
        f"opensips_dialogs={opensips_dialogs()} "
        f"freeswitch_channels={freeswitch_channels()} "
        f"carrier_channels={channel_count('graceful-drain-carrier')} "
        f"rtpengine_media_sockets={rtpengine_media_sockets()}"
    )


def cleanup_call():
    call = STATE.get("call")
    if not call:
        return
    api("POST", f"/v1/calls/{call['id']}/hangup")
    wait("OpenSIPS dialog cleanup", lambda: opensips_dialogs() == 0, 20)
    wait("Leamout FreeSWITCH channel cleanup", lambda: freeswitch_channels() == 0, 20)
    wait(
        "carrier channel cleanup",
        lambda: channel_count("graceful-drain-carrier") == 0,
        20,
    )
    wait("RTPengine media cleanup", lambda: rtpengine_media_sockets() == 0, 20)
    print("PASS call cleanup released SIP, application, and media state")


def main():
    provision()
    establish_call()
    cleanup_call()
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception as error:  # noqa: BLE001 - acceptance runner needs one failure surface
        print(f"FAIL graceful drain call-path acceptance: {error}", file=sys.stderr)
        sys.exit(1)
