#!/usr/bin/env python3
import random
import socket
import subprocess
import time

COMPOSE = ["docker", "compose", "-f", "deploy/compose.yaml", "-f", "tests/acceptance/self-hosted-managed/compose.yaml"]
DID = "+15551235001"


class Failure(RuntimeError):
    pass


def run(args):
    result = subprocess.run(args, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    if result.returncode:
        raise Failure(result.stdout)
    return result.stdout.strip()


def compose(*args):
    return run(COMPOSE + list(args))


def sql(statement):
    return compose("exec", "-T", "postgres", "psql", "-U", "leamout", "-d", "leamout", "-v", "ON_ERROR_STOP=1", "-Atc", statement)


def invite(timeout=6):
    call_id = f"{random.getrandbits(96):x}@self-hosted-managed"
    branch = f"z9hG4bK{random.getrandbits(64):x}"
    tag = f"{random.getrandbits(48):x}"
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind(("127.0.0.1", 0))
    sock.settimeout(timeout)
    port = sock.getsockname()[1]
    message = (
        f"INVITE sip:{DID}@managed-edge.test SIP/2.0\r\n"
        f"Via: SIP/2.0/UDP 127.0.0.1:{port};branch={branch};rport\r\n"
        f"Max-Forwards: 10\r\nFrom: <sip:+15557654321@carrier.test>;tag={tag}\r\n"
        f"To: <sip:{DID}@managed-edge.test>\r\nCall-ID: {call_id}\r\n"
        f"CSeq: 1 INVITE\r\nContact: <sip:carrier@127.0.0.1:{port}>\r\nContent-Length: 0\r\n\r\n"
    )
    sock.sendto(message.encode(), ("127.0.0.1", 5060))
    responses = []
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        try:
            response = sock.recv(65535).decode(errors="replace")
        except socket.timeout:
            break
        status_line = response.splitlines()[0]
        status = int(status_line.split()[1])
        responses.append(status_line)
        if status >= 200:
            break
        if status == 180:
            break
    return call_id, responses


def wait_for_channel(call_id):
    deadline = time.monotonic() + 8
    while time.monotonic() < deadline:
        channels = compose("exec", "-T", "freeswitch", "fs_cli", "-x", "show channels")
        if DID in channels:
            return
        time.sleep(0.2)
    raise Failure(f"FreeSWITCH did not receive attached call {call_id}")


def main():
    call_id, responses = invite()
    statuses = [int(response.split()[1]) for response in responses]
    if 180 not in statuses:
        raise Failure(f"healthy attachment did not reach self-hosted runtime: responses={responses}")
    wait_for_channel(call_id)
    print("PASS managed edge forwarded the DID to the self-hosted OpenSIPS and FreeSWITCH runtime")

    sql("UPDATE runtime_attachments SET health_status='unhealthy', last_checked_at=now() WHERE deployment_id='00000000-0000-0000-0000-000000005011'")
    _, responses = invite()
    statuses = [int(response.split()[1]) for response in responses]
    if not statuses or statuses[-1] != 404:
        raise Failure(f"unhealthy attachment did not fail closed: responses={responses}")
    print("PASS unhealthy runtime attachment failed closed at the managed edge")


if __name__ == "__main__":
    main()
