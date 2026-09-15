#!/bin/sh
set -eu

python -m uvicorn source.API:api \
    --app-dir /app/Spam_detection \
    --host 127.0.0.1 \
    --port 8000 &
python_pid=$!
go_pid=""

cleanup() {
    if [ -n "$go_pid" ]; then
        kill "$go_pid" 2>/dev/null || true
    fi
    kill "$python_pid" 2>/dev/null || true
}

trap cleanup INT TERM EXIT

attempt=0
until python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/health', timeout=1).read()" >/dev/null 2>&1; do
    attempt=$((attempt + 1))

    if ! kill -0 "$python_pid" 2>/dev/null; then
        echo "Python ML service exited before becoming ready" >&2
        exit 1
    fi

    if [ "$attempt" -ge 60 ]; then
        echo "Python ML service did not become ready within 60 seconds" >&2
        exit 1
    fi

    sleep 1
done

./spam-web &
go_pid=$!
wait "$go_pid"
