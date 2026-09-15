FROM golang:1.25.7-alpine AS go-builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build \
    -buildvcs=false \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/spam-web \
    ./cmd/server


FROM python:3.14.4-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    ML_API_URL=http://127.0.0.1:8000

WORKDIR /app

COPY Spam_detection/requirements-render.txt /tmp/requirements-render.txt
RUN python -m pip install --no-cache-dir -r /tmp/requirements-render.txt

COPY Spam_detection/source ./Spam_detection/source
COPY Spam_detection/artifact ./Spam_detection/artifact
COPY templates ./templates
COPY --from=go-builder /out/spam-web ./spam-web
COPY scripts/start-render.sh ./start-render.sh

RUN chmod +x ./start-render.sh ./spam-web

EXPOSE 10000

CMD ["./start-render.sh"]
