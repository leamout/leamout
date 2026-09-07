#!/usr/bin/env python3
import json
import os
import subprocess
import time
import urllib.error
import urllib.request

API = os.getenv("CLOUD_MANAGED_API_BASE", "http://127.0.0.1:8080")
PROVIDER = os.getenv("CLOUD_MANAGED_PROVIDER_STATE", "http://127.0.0.1:18090/__state")
TOKEN = os.getenv("CLOUD_MANAGED_TOKEN", "lm_org_v1smoke0_v1smoke0abcdefghijklmnopqrstuvwx")
DID = "+15551236001"
COMPOSE = ["docker", "compose", "-f", "deploy/compose.yaml", "-f", "tests/acceptance/cloud-managed/compose.yaml"]


class Failure(RuntimeError):
    pass


def topology():
    result = subprocess.run(COMPOSE + ["ps", "--status", "running", "--services"], text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    if result.returncode:
        raise Failure(result.stdout)
    required = {"postgres", "redis", "nats", "rtpengine", "freeswitch", "opensips", "server", "worker", "cloud-managed-provider"}
    missing = required - set(result.stdout.splitlines())
    if missing:
        raise Failure("cloud-managed topology is missing: " + ", ".join(sorted(missing)))


def api(method, path, payload=None, expected=(200,)):
    data = None if payload is None else json.dumps(payload).encode()
    headers = {"Accept": "application/json", "Authorization": f"Bearer {TOKEN}"}
    if data is not None:
        headers["Content-Type"] = "application/json"
        headers["Idempotency-Key"] = "cloud-managed-number-purchase"
    request = urllib.request.Request(API + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            status, raw = response.status, response.read()
    except urllib.error.HTTPError as error:
        status, raw = error.code, error.read()
    if status not in expected:
        raise Failure(f"{method} {path}: HTTP {status}: {raw.decode(errors='replace')}")
    body = json.loads(raw) if raw else None
    return body["data"] if isinstance(body, dict) and body.get("success") is True else body


def wait_for(description, probe, timeout=35):
    deadline = time.monotonic() + timeout
    last = None
    while time.monotonic() < deadline:
        try:
            value = probe()
            if value:
                return value
        except Exception as error:
            last = error
        time.sleep(0.25)
    raise Failure(f"timed out waiting for {description}: {last}")


def main():
    topology()
    print("PASS Leamout-operated Cloud + Managed runtime topology is healthy")
    available = api("GET", "/v1/numbers/available?country_code=US&contains=5551236")["numbers"]
    if len(available) != 1 or available[0]["number"] != DID:
        raise Failure(f"unexpected managed inventory: {available}")
    selection = available[0]["selection_id"]
    if not selection.startswith("sel_"):
        raise Failure("inventory did not return an opaque selection handle")
    print("PASS provider inventory is exposed as an opaque managed-number selection")

    number = api("POST", "/v1/numbers/", {"type": "managed", "selection_id": selection}, (201,))
    if number["status"] != "provisioning" or number["type"] != "managed":
        raise Failure(f"managed purchase did not persist provisioning intent: {number}")
    if number.get("carrier_connection_id") is not None:
        raise Failure("customer response exposed the platform carrier connection")

    def active_number():
        current = api("GET", f"/v1/numbers/{number['id']}")
        return current if current["status"] == "active" else None

    active = wait_for("provider operation completion", active_number)
    if active["number"] != DID or not active["voice_enabled"]:
        raise Failure(f"activated managed number is invalid: {active}")
    if "provider_id" in active or "provider_resource_id" in active or active.get("carrier_connection_id") is not None:
        raise Failure("customer response leaked internal provider or carrier state")

    provider = json.load(urllib.request.urlopen(PROVIDER, timeout=5))
    if provider["order_posts"] != 1 or provider["routing_patches"] != 1:
        raise Failure(f"provider order/routing workflow was not exactly-once: {provider}")
    if provider["voice_in_trunk_id"] != "voice-in-cloud-1":
        raise Failure(f"DID was not routed to cloud managed ingress: {provider}")
    print("PASS managed DID purchase completed and converged provider routing")
    print("PASS active customer response hides provider and platform routing identifiers")


if __name__ == "__main__":
    main()
