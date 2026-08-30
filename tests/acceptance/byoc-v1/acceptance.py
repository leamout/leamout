#!/usr/bin/env python3
import json, os, subprocess, sys, time, urllib.error, urllib.request

API = os.getenv("BYOC_V1_API_BASE", "http://127.0.0.1:8080")
TOKEN = os.getenv("BYOC_V1_TOKEN", "lm_org_v1smoke0_v1smoke0abcdefghijklmnopqrstuvwx")
ESL_PASSWORD = os.getenv("FREESWITCH_ESL_PASSWORD", "byoc-v1-esl-secret")
DID, CALLER = "+15551234567", "+15557654321"
COMPOSE = ["docker", "compose", "-f", "deploy/compose.yaml", "-f", "tests/acceptance/byoc-v1/compose.yaml"]
S, RESULTS = {}, []

class Failure(RuntimeError): pass

def run(args, check=True):
    p = subprocess.run(args, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    if check and p.returncode: raise Failure(f"command failed: {' '.join(args)}\n{p.stdout}")
    return p.stdout.strip()

def compose(*args, check=True): return run(COMPOSE + list(args), check)
def fs(command): return compose("exec", "-T", "byoc-v1-carrier", "fs_cli", "-H", "byoc-v1-carrier", "-P", "8021", "-p", ESL_PASSWORD, "-x", command)
def psql(sql): return compose("exec", "-T", "postgres", "psql", "-U", "leamout", "-d", "leamout", "-Atc", sql)

def api(method, path, payload=None, expected=(200,)):
    data = json.dumps(payload).encode() if payload is not None else None
    headers = {"Accept": "application/json", "Authorization": f"Bearer {TOKEN}"}
    if data: headers["Content-Type"] = "application/json"
    try:
        with urllib.request.urlopen(urllib.request.Request(API + path, data=data, headers=headers, method=method), timeout=15) as r:
            status, raw = r.status, r.read()
    except urllib.error.HTTPError as e: status, raw = e.code, e.read()
    if status not in expected: raise Failure(f"{method} {path}: HTTP {status}, body={raw.decode(errors='replace')}")
    if not raw: return None
    body = json.loads(raw)
    return body["data"] if body.get("success") is True else body

def wait(description, probe, timeout=20):
    end = time.monotonic() + timeout
    while time.monotonic() < end:
        try:
            value = probe()
            if value: return value
        except Exception: pass
        time.sleep(.25)
    raise Failure(f"timed out waiting for {description}")

def check(name, fn):
    try:
        detail = fn() or ""
        RESULTS.append(True); print(f"PASS {len(RESULTS):02d} {name}: {detail}")
    except Exception as e:
        RESULTS.append(False); print(f"FAIL {len(RESULTS):02d} {name}: {e}")

def deploy():
    wait("API readiness", lambda: urllib.request.urlopen(API + "/readyz", timeout=2).status == 204, 45)
    required = {"postgres", "redis", "nats", "rtpengine", "freeswitch", "opensips", "api", "worker", "byoc-v1-carrier"}
    missing = required - set(compose("ps", "--status", "running", "--services").splitlines())
    if missing: raise Failure("missing services: " + ", ".join(sorted(missing)))
    return "signaling and control stack is running"

def provider():
    body = api("GET", "/v1/carrier-providers/")
    item = next((x for x in body["carrier_providers"] if x["slug"] == "generic-sip"), None)
    if not item or item["adapter"] != "sip" or item["status"] != "active": raise Failure("generic provider is missing or invalid")
    S["provider"] = item; return f"production provider {item['id']} is active"

def connection_and_auth():
    item = api("POST", "/v1/carrier-connections/", {"provider_id": S["provider"]["id"], "name": "BYOC synthetic carrier", "inbound_enabled": True}, (201,))
    S["connection"] = item
    item = api("PUT", f"/v1/carrier-connections/{item['id']}/outbound-auth", {"method":"digest", "username":"byoc-user", "secret":"first-secret"})
    if not item["has_outbound_credentials"]: raise Failure("outbound credentials were not marked present")
    api("PUT", f"/v1/carrier-connections/{item['id']}/outbound-auth", {"method":"digest", "username":"byoc-user", "secret":"rotated-secret"})
    stored = psql(f"SELECT auth_secret_ciphertext FROM carrier_connections WHERE id='{item['id']}'")
    if not stored or "rotated-secret" in stored: raise Failure("credential was not encrypted")
    api("DELETE", f"/v1/carrier-connections/{item['id']}/outbound-auth", expected=(204,))
    return "digest credential set, rotated, encrypted, and removed"

def trunk():
    item = api("POST", "/v1/trunks/", {"carrier_connection_id":S["connection"]["id"], "name":"BYOC trunk", "direction":"bidirectional"}, (201,)); S["trunk"] = item
    endpoint = api("POST", f"/v1/trunks/{item['id']}/endpoints", {"host":"byoc-v1-carrier", "port":5060, "transport":"udp", "direction":"bidirectional", "priority":10}, (201,))
    S["endpoint"] = endpoint; return f"trunk {item['id']} routes to {endpoint['host']}"

def number_and_app():
    number = api("POST", "/v1/numbers/", {"number":DID, "country_code":"US", "voice_enabled":True}, (201,)); S["number"] = number
    assigned = api("PUT", f"/v1/numbers/{number['id']}/carrier-connection", {"carrier_connection_id":S["connection"]["id"]})
    if assigned.get("carrier_connection_id") != S["connection"]["id"]: raise Failure("DID ownership was not assigned")
    app = api("POST", "/v1/voice-applications/", {"name":"BYOC ingress", "caller_id":CALLER}, (201,)); S["app"] = app
    api("POST", f"/v1/voice-applications/{app['id']}/bindings", {"phone_number_id":number["id"]}, (201,))
    return "DID ownership and application binding use public APIs"

def rejected_before_allowlist():
    output = fs(f"originate {{origination_caller_id_number={CALLER}}}sofia/internal/{DID}@opensips:5060 &park()")
    if "+OK" in output: raise Failure("untrusted carrier source was accepted")
    return "unknown carrier source is rejected"

def add_source_and_inbound():
    source = api("POST", f"/v1/carrier-connections/{S['connection']['id']}/source-ips", {"cidr":"172.30.0.50/32"}, (201,))
    before = {x["id"] for x in api("GET", "/v1/calls/?limit=100")["calls"]}
    output = fs(f"originate {{origination_caller_id_number={CALLER}}}sofia/internal/{DID}@opensips:5060 &park()")
    if "+OK" not in output: raise Failure("allowlisted carrier was rejected: " + output)
    call = wait("inbound call", lambda: next((x for x in api("GET", "/v1/calls/?limit=100")["calls"] if x["id"] not in before and x["direction"] == "inbound"), None))
    S["inbound"] = call
    api("DELETE", f"/v1/carrier-connections/{S['connection']['id']}/source-ips/{source['id']}", expected=(204,))
    rejected = fs(f"originate {{origination_caller_id_number={CALLER}}}sofia/internal/{DID}@opensips:5060 &park()")
    if "+OK" in rejected: raise Failure("removed carrier source remained authorized")
    return "source-IP addition and removal took effect without OpenSIPS restart"

def outbound():
    call = api("POST", "/v1/calls/", {"application_id":S["app"]["id"], "trunk_id":S["trunk"]["id"], "from":CALLER, "to":DID}, (201,)); S["outbound"] = call
    if call.get("trunk_id") != S["trunk"]["id"] or not call.get("sip_call_id"): raise Failure("outbound route attribution is incomplete")
    return f"outbound call {call['id']} reached the BYOC carrier"

def disable_rejects_routes():
    api("PATCH", f"/v1/carrier-connections/{S['connection']['id']}", {"status":"disabled"})
    api("POST", "/v1/calls/", {"trunk_id":S["trunk"]["id"], "from":CALLER, "to":DID}, (409,))
    return "disabled carrier connection rejects new outbound routes"

def restart_persistence():
    compose("restart", "opensips", "api")
    wait("API recovery", lambda: urllib.request.urlopen(API + "/readyz", timeout=2).status == 204, 45)
    item = api("GET", f"/v1/carrier-connections/{S['connection']['id']}")
    if item["status"] != "disabled": raise Failure("carrier state changed after restart")
    return "carrier configuration survives OpenSIPS and API restart"

def main():
    for name, fn in [("Deploy BYOC stack",deploy),("Discover generic SIP provider",provider),("Manage encrypted carrier auth",connection_and_auth),("Provision trunk endpoint",trunk),("Assign DID ownership",number_and_app),("Reject unknown source",rejected_before_allowlist),("Apply source IP live",add_source_and_inbound),("Complete outbound route",outbound),("Disable carrier routing",disable_rejects_routes),("Recover configuration",restart_persistence)]: check(name, fn)
    failed = len([x for x in RESULTS if not x]); print(f"\nBYOC v1 acceptance {'FAILED' if failed else 'PASSED'}: {failed} failure(s).")
    return 1 if failed else 0

if __name__ == "__main__": sys.exit(main())
