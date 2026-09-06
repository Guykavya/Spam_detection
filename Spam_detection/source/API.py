from pathlib import Path
import joblib
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel


api = FastAPI(title="Email spam detection API")

MODEL_PATH = (
    Path(__file__).resolve().parent.parent
    /"artifact"
    /"spam_detection_pipline.joblib"
)

pipeline = joblib.load(MODEL_PATH)

class EmailRequest(BaseModel):
    subject: str
    body: str


@api.get("/health")
def health():
    return {"status": "ok"}

@api.post("/predict")
def predict_email(email: EmailRequest):
    if not email.subject.strip() and not email.body.strip():
        raise HTTPException(
            status_code=422,
            detail="Enter an email subject or body"
        )

    email_text = email.subject + "\n" + email.body

    label = int(pipeline.predict([email_text])[0])

    return{
        "label":label,
        "prediction": "spam" if label == 1 else "not-spam"
    }

