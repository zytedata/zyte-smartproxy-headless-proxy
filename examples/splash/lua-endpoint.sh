#!/bin/sh
set -eu

URL="http://books.toscrape.com"
PROXY_HOST="proxy"
PROXY_PORT="3128"
CURRENT_DIR="$(dirname "$0")"

curl -ifsL \
    --compressed \
    -X POST \
    -H "Content-Type: application/json" \
    -d "$(python3 -c "import json, os.path; print(json.dumps({'url': '${URL}', 'lua_source': open(os.path.join('${CURRENT_DIR}', 'example.lua')).read(), 'proxy_host': '${PROXY_HOST}', 'proxy_port': '${PROXY_PORT}'}))")" \
    "http://localhost:8050/execute"
