#!/usr/bin/env python3
import json
import random
import socket
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

SIGNALING_IP = "172.30.0.60"
state = {"outbound_invites": 0, "last_destination": "", "last_call_id": "", "internal_route_header_seen": False}
inbound_sockets = []


def headers(message):
    result = {}
    for line in message.splitlines()[1:]:
        if ":" in line:
            key, value = line.split(":", 1)
            result.setdefault(key.lower(), value.strip())
    return result


def response(status, reason, request_headers):
    return (
        f"SIP/2.0 {status} {reason}\r\n"
        f"Via: {request_headers.get('via', '')}\r\n"
        f"From: {request_headers.get('from', '')}\r\n"
        f"To: {request_headers.get('to', '')};tag=cloud-wholesale\r\n"
        f"Call-ID: {request_headers.get('call-id', '')}\r\n"
        f"CSeq: {request_headers.get('cseq', '')}\r\nContent-Length: 0\r\n\r\n"
    ).encode()


def sip_server():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind((SIGNALING_IP, 5060))
    while True:
        data, peer = sock.recvfrom(65535)
        message = data.decode(errors="replace")
        request_headers = headers(message)
        if message.startswith("OPTIONS "):
            sock.sendto(response(200, "OK", request_headers), peer)
        elif message.startswith("INVITE "):
            state["outbound_invites"] += 1
            state["last_destination"] = message.split()[1]
            state["last_call_id"] = request_headers.get("call-id", "")
            state["internal_route_header_seen"] |= "\nX-Leamout-Route-URI:" in "\n" + message
            sock.sendto(response(200, "OK", request_headers), peer)


def originate_inbound():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind((SIGNALING_IP, 0))
    sock.settimeout(8)
    inbound_sockets.append(sock)
    call_id = f"{random.getrandbits(96):x}@cloud-managed-wholesale"
    port = sock.getsockname()[1]
    sdp = (
        "v=0\r\n"
        f"o=- 1 1 IN IP4 {SIGNALING_IP}\r\n"
        "s=cloud-managed-acceptance\r\n"
        f"c=IN IP4 {SIGNALING_IP}\r\n"
        "t=0 0\r\n"
        "m=audio 40000 RTP/AVP 0 101\r\n"
        "a=rtpmap:0 PCMU/8000\r\n"
        "a=rtpmap:101 telephone-event/8000\r\n"
        "a=fmtp:101 0-16\r\n"
        "a=sendrecv\r\n"
    )
    message = (
        "INVITE sip:+15551236001@cloud-managed.local SIP/2.0\r\n"
        f"Via: SIP/2.0/UDP cloud-managed-wholesale:{port};branch=z9hG4bK{random.getrandbits(64):x};rport\r\n"
        "Max-Forwards: 10\r\nFrom: <sip:+15557654321@wholesale.test>;tag=inbound\r\n"
        "To: <sip:+15551236001@cloud-managed.local>\r\n"
        f"Call-ID: {call_id}\r\nCSeq: 1 INVITE\r\n"
        f"Contact: <sip:carrier@cloud-managed-wholesale:{port}>\r\n"
        f"Content-Type: application/sdp\r\nContent-Length: {len(sdp.encode())}\r\n\r\n{sdp}"
    )
    # Use the cloud edge's public-signaling address so carrier source-IP
    # authentication observes this simulator's fixed public-signaling IP.
    sock.sendto(message.encode(), ("172.30.0.10", 5060))
    responses = []
    deadline = time.monotonic() + 8
    while time.monotonic() < deadline:
        try:
            reply = sock.recv(65535).decode(errors="replace")
        except socket.timeout:
            break
        status_line = reply.splitlines()[0]
        status = int(status_line.split()[1])
        responses.append(status_line)
        if status == 180 or status >= 200:
            break
    return {
        "call_id": call_id,
        "statuses": [int(item.split()[1]) for item in responses],
        "responses": responses,
    }


class Status(BaseHTTPRequestHandler):
    def reply(self, status, payload):
        body = json.dumps(payload).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        self.reply(200, state)

    def do_POST(self):
        if self.path != "/originate":
            self.reply(404, {"error": "not found"})
            return
        self.reply(200, originate_inbound())

    def log_message(self, *_):
        pass


threading.Thread(target=sip_server, daemon=True).start()
ThreadingHTTPServer(("0.0.0.0", 8088), Status).serve_forever()
