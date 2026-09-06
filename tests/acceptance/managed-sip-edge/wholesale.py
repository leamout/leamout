#!/usr/bin/env python3
import json, socket, threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

state = {"invites": 0, "proxy_authorization_seen": False}

def sip():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind(("0.0.0.0", 5060))
    while True:
        data, peer = sock.recvfrom(65535)
        text = data.decode(errors="replace")
        if not text.startswith("INVITE "):
            continue
        state["invites"] += 1
        state["proxy_authorization_seen"] |= "\nProxy-Authorization:" in "\n" + text
        headers = {}
        for line in text.splitlines()[1:]:
            if ":" in line:
                key, value = line.split(":", 1)
                headers.setdefault(key.lower(), value.strip())
        response = (
            "SIP/2.0 200 OK\r\n"
            f"Via: {headers.get('via', '')}\r\n"
            f"From: {headers.get('from', '')}\r\n"
            f"To: {headers.get('to', '')};tag=wholesale\r\n"
            f"Call-ID: {headers.get('call-id', '')}\r\n"
            f"CSeq: {headers.get('cseq', '')}\r\nContent-Length: 0\r\n\r\n"
        )
        sock.sendto(response.encode(), peer)

class Status(BaseHTTPRequestHandler):
    def do_GET(self):
        body = json.dumps(state).encode()
        self.send_response(200); self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body))); self.end_headers(); self.wfile.write(body)
    def log_message(self, *_): pass

threading.Thread(target=sip, daemon=True).start()
ThreadingHTTPServer(("0.0.0.0", 8088), Status).serve_forever()
