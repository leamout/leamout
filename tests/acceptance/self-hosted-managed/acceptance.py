#!/usr/bin/env python3
import os
import random
import socket
import subprocess
import time

COMPOSE = ["docker", "compose", "-f", "deploy/compose.yaml", "-f", "tests/acceptance/self-hosted-managed/compose.yaml"]
DID = "+15551235001"
ESL_PASSWORD = os.environ["FREESWITCH_ESL_PASSWORD"]


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
    sdp = (
        "v=0\r\n"
        "o=- 1 1 IN IP4 127.0.0.1\r\n"
        "s=leamout-managed-carrier-acceptance\r\n"
        "c=IN IP4 127.0.0.1\r\n"
        "t=0 0\r\n"
        "m=audio 40000 RTP/AVP 0 101\r\n"
        "a=rtpmap:0 PCMU/8000\r\n"
        "a=rtpmap:101 telephone-event/8000\r\n"
        "a=fmtp:101 0-16\r\n"
        "a=sendrecv\r\n"
    )
    message = (
        f"INVITE sip:{DID}@self-hosted.test SIP/2.0\r\n"
        f"Via: SIP/2.0/UDP 127.0.0.1:{port};branch={branch};rport\r\n"
        "Max-Forwards: 10\r\n"
        f"From: <sip:+15557654321@sip.leamout.com>;tag={tag}\r\n"
        f"To: <sip:{DID}@self-hosted.test>\r\n"
        f"Call-ID: {call_id}\r\n"
        "CSeq: 1 INVITE\r\n"
        f"Contact: <sip:carrier@127.0.0.1:{port}>\r\n"
        "Content-Type: application/sdp\r\n"
        f"Content-Length: {len(sdp.encode())}\r\n\r\n{sdp}"
    )
    sock.sendto(message.encode(), ("127.0.0.1", 5070))
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
        if status >= 200 or status == 180:
            break
    return call_id, responses


def wait_for_channel(call_id):
    deadline = time.monotonic() + 8
    while time.monotonic() < deadline:
        channels = compose(
            "exec", "-T", "freeswitch", "fs_cli",
            "-H", "127.0.0.1", "-P", "8021", "-p", ESL_PASSWORD,
            "-x", "show channels",
        )
        if DID in channels:
            return
        time.sleep(0.2)
    raise Failure(f"FreeSWITCH did not receive ordinary carrier call {call_id}")


def main():
    shape = sql(
        "SELECT cp.slug || ',' || cc.scope || ',' || pn.provisioning_mode "
        "FROM carrier_connections cc "
        "JOIN carrier_providers cp ON cp.id=cc.provider_id "
        "JOIN phone_numbers pn ON pn.carrier_connection_id=cc.id "
        "WHERE cc.id='00000000-0000-0000-0000-000000005020'::uuid"
    )
    if shape != "leamout,organization,byoc":
        raise Failure(f"self-hosted managed carrier is not ordinary SIP/BYOC state: {shape}")
    print("PASS self-hosted runtime models Leamout Managed Carrier as an ordinary SIP carrier connection")

    call_id, responses = invite()
    statuses = [int(response.split()[1]) for response in responses]
    if 180 not in statuses:
        raise Failure(f"ordinary Leamout carrier ingress did not reach self-hosted runtime: responses={responses}")
    wait_for_channel(call_id)
    print("PASS Leamout Managed Carrier reached self-hosted OpenSIPS and FreeSWITCH through normal carrier ingress")

    sql("UPDATE carrier_connections SET status='disabled' WHERE id='00000000-0000-0000-0000-000000005020'")
    try:
        _, responses = invite()
    finally:
        sql("UPDATE carrier_connections SET status='active' WHERE id='00000000-0000-0000-0000-000000005020'")
    statuses = [int(response.split()[1]) for response in responses]
    if 180 in statuses:
        raise Failure(f"disabled ordinary carrier connection still reached the self-hosted runtime: responses={responses}")
    print("PASS self-hosted managed-carrier ingress is controlled by the generic carrier connection state")


if __name__ == "__main__":
    main()
