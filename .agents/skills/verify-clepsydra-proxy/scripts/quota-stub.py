import json
import sys
from datetime import datetime, timedelta, timezone
from http.server import BaseHTTPRequestHandler, HTTPServer

port = int(sys.argv[1])
status = int(sys.argv[2]) if len(sys.argv) > 2 else 200
log_path = sys.argv[3] if len(sys.argv) > 3 else None


def reset(hours):
    return (datetime.now(timezone.utc) + timedelta(hours=hours)).strftime("%Y-%m-%dT%H:%M:%SZ")


def payload():
    return {
        "groups": [
            {"displayName": "Gemini models", "buckets": [
                {"bucketId": "gemini-5h", "window": "5h", "remainingFraction": 0.75, "resetTime": reset(3)},
                {"bucketId": "gemini-weekly", "window": "weekly", "remainingFraction": 0.9, "resetTime": reset(100)},
            ]},
            {"displayName": "Claude and GPT models", "buckets": [
                {"bucketId": "3p-5h", "window": "5h", "remainingFraction": 1, "resetTime": reset(5)},
                {"bucketId": "3p-weekly", "window": "weekly", "remainingFraction": 0.5, "resetTime": reset(120)},
                {"bucketId": "3p-disabled", "window": "weekly", "remainingFraction": 1, "disabled": True},
            ]},
        ]
    }


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("Content-Length") or 0)
        body = self.rfile.read(length).decode()
        if not self.path.endswith(":retrieveUserQuotaSummary"):
            self.send_response(404)
            self.send_header("Content-Length", "0")
            self.end_headers()
            return
        if log_path:
            with open(log_path, "a") as log:
                log.write(json.dumps({"path": self.path, "body": body, "ua": self.headers.get("User-Agent")}) + "\n")
        data = json.dumps(payload() if status == 200 else {"error": {"code": status}}).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def log_message(self, *args):
        pass


HTTPServer(("127.0.0.1", port), Handler).serve_forever()
