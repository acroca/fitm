FROM golang:1.22-bullseye as build
WORKDIR /build
COPY go.mod go.sum /build/
RUN go mod download
COPY cmd cmd
COPY internal internal
RUN go build -o fitm ./cmd/fitm/main.go

FROM mitmproxy/mitmproxy:10.2.2
WORKDIR /app
COPY fitm.py .
COPY --from=build /build/fitm .
CMD ["./fitm", "run"]
