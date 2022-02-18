#!/usr/bin/env bash

set -e
function require_tool {
  if ! command -v "$1" &> /dev/null
  then
      echo "$1 could not be found"
      exit 1
  fi
}
require_tool "git"
require_tool "docker"
require_tool "curl"

tmp_dir=$(mktemp -d)
mkdir -p $HOME/.config/fitm/dotmitmproxy

function cleanup {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT
cd "$tmp_dir"

echo "Cloning repository..."
git clone -q git@github.com:acroca/fitm.git
cd fitm

echo "Starting FITM..."

docker compose up -d 2> /dev/null

curl --retry 30 --retry-delay 1 -s -o /dev/null localhost:4000

curl -s -o /dev/null -X POST http://127.0.0.1:4000/buckets -d '{"id":"default"}'
curl -s -o /dev/null -X POST http://127.0.0.1:4000/users -d '{"id":"admin","tokens":["admin"],"buckets":["default"]}'

echo "FITM running! 🎉"
echo ""
echo "Next steps:"
echo "1. Install the FITM CA Cert located here: $HOME/.config/fitm/dotmitmproxy/mitmproxy-ca-cert.pem (or p12 in Windows)"
echo "2. Configure your HTTP PROXY with the following values"
echo "     Host: 127.0.0.1"
echo "     Port: 8080"
echo "     Username: default"
echo "     Password: admin"
