# Python ML Service

This directory contains the FastAPI prediction service, saved TF-IDF + Logistic Regression pipeline, model-development notebooks, and API tests.

From the repository root, start it with:

```bash
python -m uvicorn source.API:api \
  --app-dir Spam_detection \
  --host 127.0.0.1 \
  --port 8000 \
  --reload
```

Use `requirements-render.txt` for the API/runtime dependencies. The older `requirement.txt` is the complete notebook and development environment snapshot.

See the repository's main `README.md` for local Go integration, Google OAuth configuration, testing, Docker, and Render deployment instructions.
