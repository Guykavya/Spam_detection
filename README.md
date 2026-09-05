# Spam_detection
# Email Spam Detection

College project using Python machine learning and a Go web application.

## Current status

- Dataset cleaning and EDA completed
- Naive Bayes and Logistic Regression compared
- TF-IDF + Logistic Regression selected
- Python prediction API implemented
- Four API tests passing
- Go frontend and Gmail integration pending

## Model evaluation

Results on the dataset test split:

| Model | Accuracy |
|---|---:|
| Multinomial Naive Bayes | 97.00% |
| Logistic Regression | 98.38% |

Real-world email performance may differ from these dataset results.

## Project folders

- `Python-services/source/API.py`: FastAPI application
- `Python-services/artifact/`: saved prediction pipeline
- `Python-services/tests/`: API tests
- `Python-services/Notebooks/`: EDA and training notebooks
- `Python-services/requirement.txt`: Python dependencies
- `go-web/`: Go backend and frontend
- `presentation/`: presentation materials
- `docs/`: additional documentation

## Python setup

The project was developed using Python 3.14.4.
Use the same Python version when possible.

Run these commands from the repository root.

### Linux / macOS

```bash
python3 -m venv .venv
source .venv/bin/activate
python -m pip install -r Python-services/requirement.txt
```

### Windows PowerShell

```powershell
py -3.14 -m venv .venv
.\.venv\Scripts\python.exe -m pip install -r Python-services/requirement.txt
```

Activate the environment if PowerShell permits it:

```powershell
.\.venv\Scripts\Activate.ps1
```

Alternatively, use the environment's Python executable directly.

## Run the Python API

With the environment activated:

```bash
cd Python-services
python -m uvicorn source.API:api --reload
```

Windows alternative without activation, from the repository root:

```powershell
.\.venv\Scripts\python.exe -m uvicorn source.API:api --app-dir Python-services --reload
```

The trained pipeline must be present at the path configured in `source/API.py`.
Retraining and the original dataset are not required to run predictions.

- Health check: http://127.0.0.1:8000/health
- Interactive API documentation: http://127.0.0.1:8000/docs

Keep the server running while using the application.
Press Ctrl+C to stop it.

## API agreement

### GET /health

Successful response:

```json
{"status": "ok"}
```

### POST /predict

Send JSON with both fields present:

```json
{
  "subject": "Project meeting",
  "body": "Please bring your notebook tomorrow."
}
```

Response format:

```json
{
  "label": 0,
  "prediction": "not spam"
}
```

Allowed results:

- `0`: `"not spam"`
- `1`: `"spam"`

The model determines the actual result.

Validation:
- Subject and body must be strings.
- Either may be empty, but both cannot be blank.
- Invalid requests return HTTP 422.

Python combines the subject and body into one text input.

## Run Python tests

From `Python-services`, with the environment activated:

```bash
python -m pytest tests/test_api.py -v
```

Uvicorn does not need to be running for these tests.

Tests cover:
- Health response
- Valid prediction response
- Empty input rejection
- Missing required field rejection

## Go integration

Go should send HTTP POST requests to:

http://127.0.0.1:8000/predict

Use Content-Type: application/json and the request format above.

Both manual mode and Gmail mode use the same endpoint.
Go handles Gmail authentication, fetching messages, and the user interface.

Localhost works when Go and Python run on the same laptop.
Each developer should run their own local Python service.

Gmail predictions must not automatically move or delete messages.

## Collaboration

- Python work belongs in `Python-services`.
- Go and frontend work belongs in `go-web`.
- Pull the latest changes before starting work.
- Prefer a separate branch and a pull request for changes.
- Never commit credentials, Gmail tokens, private email downloads, or `.venv`.

The training dataset is excluded from this repository.
Dataset source and reproduction instructions will be documented separately.
