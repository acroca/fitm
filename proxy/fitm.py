import collections
import hvac
import json
import os
import weakref

from http import cookiejar
from typing import List, Tuple, Dict, Optional  # noqa

from mitmproxy import http, flowfilter, ctx, exceptions, connection
from mitmproxy.net.http import cookies


def domain_match(a: str, b: str) -> bool:
    if cookiejar.domain_match(a, b):  # type: ignore
        return True
    elif cookiejar.domain_match(a, b.strip(".")):  # type: ignore
        return True
    return False

class FITM:
    def __init__(self):
        self.clients = {}

    def vault_client(self, flow: http.HTTPFlow) -> hvac.Client:
        _, token = flow.metadata["proxyauth"]
        if token in self.clients:
            return self.clients[token]
        client = hvac.Client(
            url=os.environ.get('VAULT_ADDRESS', 'http://localhost:8200'),
            token=token,
        )
        if not client.is_authenticated():
            flow.response = http.Response.make(407)
            return None

        self.clients[token] = client
        return client

    def response(self, flow: http.HTTPFlow):
        assert flow.response

        bucket, token = flow.metadata["proxyauth"]
        v = self.vault_client(flow)
        if v == None:
            return False

        read_response = v.secrets.kv.read_secret_version(path='buckets/'+bucket)
        vault_cookies = json.loads(read_response['data']['data']['cookies'])

        for key, (value, attrs) in flow.response.cookies.items(multi=True):
            # FIXME: We now know that Cookie.py screws up some cookies with
            # valid RFC 822/1123 datetime specifications for expiry. Sigh.
            domain = flow.request.host
            port = flow.request.port
            path = "/"
            if "domain" in attrs:
                domain = attrs["domain"]
            if "path" in attrs:
                path = attrs["path"]

            if domain_match(flow.request.host, domain):
                new_cookies = [
                    c for c in vault_cookies
                    if not (
                        c['domain'] == domain and
                        c['port'] == port and
                        c['path'] == path and
                        c['key'] == key
                    )
                ]
                vault_cookies = new_cookies

                if not cookies.is_expired(attrs):
                    vault_cookies.append({
                        'domain': domain,
                        'port': port,
                        'path': path,
                        'key': key,
                        'value': value,
                })
        create_response = v.secrets.kv.v2.create_or_update_secret(
            path='buckets/'+bucket,
            secret=dict(cookies=json.dumps(vault_cookies)),
        )

        flow.response.cookies = []

    def request(self, flow: http.HTTPFlow):
        bucket, token = flow.metadata["proxyauth"]
        v = self.vault_client(flow)
        if v == None:
            return False

        cookie_list: List[Tuple[str, str]] = []

        read_response = v.secrets.kv.read_secret_version(path='buckets/'+bucket)
        vault_cookies = json.loads(read_response['data']['data']['cookies'])

        for cookie in vault_cookies:
            if flow.request.path.startswith(cookie["path"]) and domain_match(flow.request.host, cookie["domain"]):
                cookie_list.extend([ [cookie["key"], cookie["value"]] ])

        if cookie_list:
            flow.request.headers["cookie"] = cookies.format_cookie_header(cookie_list)

addons = [
  FITM(),
]
