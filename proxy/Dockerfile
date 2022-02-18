FROM mitmproxy/mitmproxy:7.0.4

WORKDIR /root

COPY fitm.py fitm.py

RUN pip install hvac

CMD mitmdump -s fitm.py --proxyauth any
