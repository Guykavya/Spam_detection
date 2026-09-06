package mlclient

import "strings"

// PredictRequest matches the JSON payload sent to Python
type PredictRequest struct {
	MessageID string `json:"message_id,omitempty"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

// PredictResponse matches Python's JSON output
type PredictResponse struct {
	MessageID  string `json:"message_id,omitempty"`
	Label      int    `json:"label"`      // 0 for not_spam, 1 for spam
	Prediction string `json:"prediction"` // "not spam" or "spam"
}

// BatchPredictRequest holds multiple emails for batch calls
type BatchPredictRequest struct {
	Emails []PredictRequest `json:"emails"`
}

// BatchPredictResponse holds prediction results for batch calls
type BatchPredictResponse struct {
	Results []PredictResponse `json:"results"`
}

// MockPredict generates realistic predictions for offline testing
func MockPredict(req PredictRequest) PredictResponse {
	text := strings.ToLower(req.Subject + " " + req.Body)
	isSpam := strings.Contains(text, "prize") ||
		strings.Contains(text, "urgent") ||
		strings.Contains(text, "claim") ||
		strings.Contains(text, "free") ||
		strings.Contains(text, "winner") ||
		strings.Contains(text, "account suspended")

	label := 0
	prediction := "not spam"

	if isSpam {
		label = 1
		prediction = "spam"
	}

	return PredictResponse{
		MessageID:  req.MessageID,
		Label:      label,
		Prediction: prediction,
	}
}