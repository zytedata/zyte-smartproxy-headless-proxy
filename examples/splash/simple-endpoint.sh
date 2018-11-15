#!/bin/sh
set -eu

URL="http://books.toscrape.com"
PROXY_URL="proxy:3128"

curl -ifsL \
    --compressed \
    -X POST \
    -H "Content-Type: application/json" \
    -d "$(python -c "import json; print(json.dumps({'url': '${URL}', 'proxy': 'http://${PROXY_URL}'}))")" \
    "http://localhost:8050/render.html"
