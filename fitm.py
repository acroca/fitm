import json
import os
from http import cookiejar
from mitmproxy import http
from mitmproxy.net.http import cookies

def domain_match(a: str, b: str) -> bool:
    if cookiejar.domain_match(a, b):  # type: ignore
        return True
    elif cookiejar.domain_match(a, b.strip(".")):  # type: ignore
        return True
    return False

class FITM:
    def __init__(self):
        self.cookies = []
        self.cookies_file = "/cookies/cookies.json"
        self.load_cookies()

    def response(self, flow: http.HTTPFlow) -> None:
        if "set-cookie" in flow.response.headers:
            for key, (value, attrs) in flow.response.cookies.items(multi=True):
                domain = flow.request.host
                port = flow.request.port
                path = "/"
                if "domain" in attrs:
                    domain = attrs["domain"]
                if "path" in attrs:
                    path = attrs["path"]

                if domain_match(flow.request.host, domain):
                    new_cookies = [
                        c for c in self.cookies
                        if not (
                            c['domain'] == domain and
                            c['port'] == port and
                            c['path'] == path and
                            c['key'] == key
                        )
                    ]
                    self.cookies = new_cookies

                    if not cookies.is_expired(attrs):
                        self.cookies.append({
                            'domain': domain,
                            'port': port,
                            'path': path,
                            'key': key,
                            'value': value,
                    })
            self.save_cookies()

        flow.response.cookies = []

    def request(self, flow: http.HTTPFlow) -> None:
        if flow.request.pretty_url == "http://fitm.local/mitmproxy-ca-cert.pem":
            with open("/.mitmproxy/mitmproxy-ca-cert.pem", "rb") as f:
                cert = f.read()

            flow.response = http.Response.make(200,cert,{"Content-Type": "application/x-x509-ca-cert"})
            return

        cookie_list: List[Tuple[str, str]] = []
        for cookie in self.cookies:
            if flow.request.path.startswith(cookie["path"]) and domain_match(flow.request.host, cookie["domain"]):
                cookie_list.extend([ [cookie["key"], cookie["value"]] ])
        if cookie_list:
            flow.request.headers["cookie"] = cookies.format_cookie_header(cookie_list)

    def save_cookies(self):
        cookies_dir = os.path.dirname(self.cookies_file)
        if not os.path.exists(cookies_dir):
            os.makedirs(cookies_dir)

        with open(self.cookies_file, "w") as f:
            json.dump(self.cookies, f)

    def load_cookies(self):
        if os.path.exists(self.cookies_file):
            with open(self.cookies_file, "r") as f:
                self.cookies = json.load(f)

addons = [
    FITM(),
]
