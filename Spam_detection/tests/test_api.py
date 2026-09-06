from fastapi.testclient import TestClient
from source.API import api

client = TestClient(api)

def test_health():
    res = client.get("/health")

    assert res.status_code == 200
    assert res.json() == {"status": "ok"}


def test_predict_valid_email():
    response = client.post(
        "/predict",
        json={
            "subject": "Project meeting",
            "body": "Please bring your notebook tomorrow."
        }
    )

    assert response.status_code == 200

    data = response.json()

    assert data["label"] in [0, 1]

    expected_prediction = (
        "spam" if data["label"] == 1 else "not-spam"
    )

    assert data["prediction"] == expected_prediction


def test_predict_empty_email():
    response = client.post(
        "/predict",
        json={"subject": "   ", "body": ""}
    )

    assert response.status_code == 422
    assert response.json()["detail"] == "Enter an email subject or body"


def test_predict_missing_body():
    response = client.post(
        "/predict",
        json={"subject": "Project meeting"}
    )

    assert response.status_code == 422

    errors = response.json()["detail"]

    assert any(
        error["loc"] == ["body", "body"]
        and error["type"] == "missing"
        for error in errors
    )