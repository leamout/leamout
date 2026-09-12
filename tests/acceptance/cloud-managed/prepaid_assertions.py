#!/usr/bin/env python3
import json
import os
import subprocess
import urllib.error
import urllib.request

API = os.getenv("CLOUD_MANAGED_API_BASE", "http://127.0.0.1:8080")
PROVIDER = os.getenv("CLOUD_MANAGED_PROVIDER_STATE", "http://127.0.0.1:18090/__state")
TOKEN = os.getenv("CLOUD_MANAGED_TOKEN", "lm_org_v1smoke0_v1smoke0abcdefghijklmnopqrstuvwx")
COMPOSE = ["docker", "compose", "-f", "deploy/compose.yaml", "-f", "tests/acceptance/cloud-managed/compose.yaml"]


class Failure(RuntimeError):
    pass


def api(method, path, payload=None, expected=(200,)):
    data = None if payload is None else json.dumps(payload).encode()
    headers = {"Accept": "application/json", "Authorization": f"Bearer {TOKEN}"}
    if data is not None:
        headers["Content-Type"] = "application/json"
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


def sql(statement):
    result = subprocess.run(
        COMPOSE + ["exec", "-T", "postgres", "psql", "-U", "leamout", "-d", "leamout", "-Atc", statement],
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )
    if result.returncode:
        raise Failure(result.stdout)
    return result.stdout.strip()


def provider_state():
    return json.load(urllib.request.urlopen(PROVIDER, timeout=5))


def main():
    available = api("GET", "/v1/numbers/available?country_code=US&contains=5551236")["numbers"]
    if len(available) != 1:
        raise Failure(f"expected one managed number quote, got {available}")
    quote = available[0].get("price")
    if quote != {"amount_minor": 2500, "currency": "USD"}:
        raise Failure(f"unexpected Leamout managed-number quote: {quote}")
    print("PASS managed number search exposes the Leamout customer quote")

    reservation = sql(
        "SELECT status || ',' || amount_minor::text || ',' || COALESCE(captured_amount_minor, 0)::text "
        "FROM wallet_reservations "
        "WHERE organization_id='00000000-0000-0000-0000-000000006001'::uuid "
        "AND operation_type='managed_number_purchase'"
    )
    if reservation != "captured,2500,2500":
        raise Failure(f"managed number reservation was not captured exactly once: {reservation!r}")

    capture = sql(
        "SELECT count(*)::text || ',' || COALESCE(sum(amount_minor), 0)::text "
        "FROM wallet_ledger_entries "
        "WHERE wallet_id='00000000-0000-0000-0000-000000006206'::uuid "
        "AND entry_type='capture'"
    )
    if capture != "1,-2500":
        raise Failure(f"managed number purchase ledger capture is invalid: {capture}")

    balance = sql(
        "SELECT "
        "COALESCE((SELECT sum(amount_minor) FROM wallet_ledger_entries WHERE wallet_id=w.id),0)::text || ',' || "
        "COALESCE((SELECT sum(amount_minor) FROM wallet_reservations WHERE wallet_id=w.id AND status='active'),0)::text "
        "FROM wallets w WHERE w.id='00000000-0000-0000-0000-000000006206'::uuid"
    )
    if balance != "7500,0":
        raise Failure(f"wallet balance/active reservation state is invalid: {balance}")
    print("PASS managed DID provider success captured one prepaid debit and left no active hold")

    before = provider_state()["order_posts"]
    sql(
        "INSERT INTO wallet_ledger_entries ("
        "wallet_id, organization_id, entry_type, amount_minor, source_type, source_id, idempotency_key, metadata"
        ") VALUES ("
        "'00000000-0000-0000-0000-000000006206'::uuid,"
        "'00000000-0000-0000-0000-000000006001'::uuid,"
        "'adjustment_debit',-6000,'acceptance_fixture','insufficient-funds',"
        "'cloud-managed-insufficient-funds','{}'::jsonb)"
    )
    available = api("GET", "/v1/numbers/available?country_code=US&contains=5551236")["numbers"]
    selection = available[0]["selection_id"]
    api("POST", "/v1/numbers/", {"type": "managed", "selection_id": selection}, expected=(409,))
    after = provider_state()["order_posts"]
    if after != before:
        raise Failure(f"insufficient funds reached DIDWW: order_posts changed from {before} to {after}")
    active_holds = sql(
        "SELECT count(*) FROM wallet_reservations "
        "WHERE organization_id='00000000-0000-0000-0000-000000006001'::uuid AND status='active'"
    )
    if active_holds != "0":
        raise Failure(f"insufficient-funds attempt left an active hold: {active_holds}")
    print("PASS insufficient prepaid funds fail before any DIDWW provider order")


if __name__ == "__main__":
    main()
