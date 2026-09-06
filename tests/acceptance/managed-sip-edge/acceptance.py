#!/usr/bin/env python3
import hashlib, json, os, random, re, socket, subprocess, sys, time, urllib.request

COMPOSE = ["docker", "compose", "-f", "deploy/compose.yaml", "-f", "tests/acceptance/managed-sip-edge/compose.yaml"]
REALM, USER, PASSWORD = "sip.leamout.com", "edge-user", "edge-password"
CALLER, FOREIGN, DESTINATION = "+15551234001", "+15551234999", "+15551234002"
ORG, TRUNK = "00000000-0000-0000-0000-000000004001", "00000000-0000-0000-0000-000000004030"

class Failure(RuntimeError): pass
def run(args):
    p = subprocess.run(args, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    if p.returncode: raise Failure(p.stdout)
    return p.stdout.strip()
def compose(*args): return run(COMPOSE + list(args))
def sql(statement): return compose("exec", "-T", "postgres", "psql", "-U", "leamout", "-d", "leamout", "-v", "ON_ERROR_STOP=1", "-Atc", statement)
def wholesale(): return json.load(urllib.request.urlopen("http://127.0.0.1:18088", timeout=2))

def digest_challenge(response):
    match = re.search(r'^Proxy-Authenticate:\s*Digest\s+(.+)$', response, re.I | re.M)
    if not match: raise Failure("407 response did not contain a Digest challenge")
    return {key: quoted or bare for key, quoted, bare in re.findall(r'(\w+)=(?:"([^"]*)"|([^,\s]+))', match.group(1))}

def new_branch(): return f"z9hG4bK{random.getrandbits(64):x}"

def invite(caller=CALLER, password=PASSWORD, authenticate=True):
    destination = ("127.0.0.1", 5060); call_id = f"{random.getrandbits(96):x}@acceptance"
    tag = f"{random.getrandbits(48):x}"
    uri = f"sip:{DESTINATION}@{REALM}"
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM); sock.bind(("127.0.0.1", 0)); sock.settimeout(5)
    port = sock.getsockname()[1]
    def message(branch, cseq, auth=""):
        return (f"INVITE {uri} SIP/2.0\r\nVia: SIP/2.0/UDP 127.0.0.1:{port};branch={branch};rport\r\n"
                f"Max-Forwards: 10\r\nFrom: <sip:{caller}@{REALM}>;tag={tag}\r\nTo: <{uri}>\r\n"
                f"Call-ID: {call_id}\r\nCSeq: {cseq} INVITE\r\nContact: <sip:test@127.0.0.1:{port}>\r\n{auth}Content-Length: 0\r\n\r\n")
    sock.sendto(message(new_branch(), 1).encode(), destination); first = sock.recv(65535).decode(errors="replace")
    if not authenticate: return int(first.split()[1]), first
    if int(first.split()[1]) != 407: raise Failure("initial INVITE was not challenged: " + first.splitlines()[0])
    challenge = digest_challenge(first); nonce = challenge["nonce"]
    ha1 = hashlib.md5(f"{USER}:{REALM}:{password}".encode()).hexdigest()
    ha2 = hashlib.md5(f"INVITE:{uri}".encode()).hexdigest()
    qop = challenge.get("qop", "").split(",")[0]
    if qop:
        nc, cnonce = "00000001", f"{random.getrandbits(64):x}"
        response = hashlib.md5(f"{ha1}:{nonce}:{nc}:{cnonce}:{qop}:{ha2}".encode()).hexdigest()
        extra = f", qop={qop}, nc={nc}, cnonce=\"{cnonce}\""
    else:
        response = hashlib.md5(f"{ha1}:{nonce}:{ha2}".encode()).hexdigest(); extra = ""
    auth = f'Proxy-Authorization: Digest username="{USER}", realm="{REALM}", nonce="{nonce}", uri="{uri}", response="{response}", algorithm=MD5{extra}\r\n'
    sock.sendto(message(new_branch(), 2, auth).encode(), destination)
    final = sock.recv(65535).decode(errors="replace"); return int(final.split()[1]), final

def assert_no_wholesale(before):
    time.sleep(.5)
    if wholesale()["invites"] != before: raise Failure("unauthorized INVITE reached wholesale")

def case(name, fn):
    try: print("PASS", name + ":", fn() or "ok"); return True
    except Exception as exc: print("FAIL", name + ":", exc); return False
    finally: time.sleep(1.1)

def no_auth():
    before = wholesale()["invites"]; status, _ = invite(authenticate=False)
    if status != 407: raise Failure(f"status={status}, want 407")
    assert_no_wholesale(before)
def wrong_password():
    before = wholesale()["invites"]; status, _ = invite(password="wrong-password")
    if status < 400: raise Failure(f"wrong password accepted with {status}")
    assert_no_wholesale(before)
def unauthorized_caller():
    before = wholesale()["invites"]; status, _ = invite(caller=FOREIGN)
    if status != 403: raise Failure(f"status={status}, want 403")
    assert_no_wholesale(before)
def authorized():
    before = wholesale()["invites"]; status, _ = invite()
    if status != 200: raise Failure(f"status={status}, want 200")
    end = time.monotonic() + 3
    while time.monotonic() < end and wholesale()["invites"] == before: time.sleep(.1)
    state = wholesale()
    if state["invites"] != before + 1: raise Failure("wholesale INVITE not observed")
    if state["proxy_authorization_seen"]: raise Failure("customer Proxy-Authorization leaked to wholesale")
def denied_while(statement):
    before = wholesale()["invites"]
    try:
        sql(statement); status, _ = invite()
        if status < 400: raise Failure(f"call accepted with {status}")
        assert_no_wholesale(before)
    finally:
        sql(f"UPDATE organizations SET status='active' WHERE id='{ORG}'; UPDATE trunks SET status='active' WHERE id='{TRUNK}'; UPDATE entitlements SET enabled=true WHERE entitlement_key='voice.managed.enabled' AND plan_id IS NOT NULL")

def main():
    tests = [
        ("valid trunk without auth is challenged", no_auth),
        ("wrong password is rejected before wholesale", wrong_password),
        ("unauthorized caller ID is forbidden", unauthorized_caller),
        ("valid credential, caller ID, and entitlement is authorized", authorized),
        ("inactive trunk fails closed", lambda: denied_while(f"UPDATE trunks SET status='disabled' WHERE id='{TRUNK}'")),
        ("inactive organization fails closed", lambda: denied_while(f"UPDATE organizations SET status='disabled' WHERE id='{ORG}'")),
        ("disabled managed entitlement is denied", lambda: denied_while("UPDATE entitlements SET enabled=false WHERE entitlement_key='voice.managed.enabled' AND plan_id IS NOT NULL")),
        ("Proxy-Authorization is stripped from wholesale", authorized),
    ]
    passed = [case(name, fn) for name, fn in tests]
    print(f"\nManaged SIP edge acceptance: {sum(passed)}/{len(passed)} passed")
    return 0 if all(passed) else 1
if __name__ == "__main__": sys.exit(main())
