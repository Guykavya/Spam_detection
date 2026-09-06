package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"spam-email-detector/go-backend/karan/internal/gmailclient"
	"spam-email-detector/go-backend/karan/internal/mlclient"
)

// GmailResult combines email metadata with the ML model's prediction output
type GmailResult struct {
	MessageID  string `json:"message_id"`
	Subject    string `json:"subject"`
	Label      int    `json:"label"`      // 0 = not_spam, 1 = spam
	Prediction string `json:"prediction"` // "not spam" or "spam"
}

func main() {
	// Initialize live Python ML Client pointing to Python service port
	pythonClient := mlclient.NewClient("http://127.0.0.1:8000")

	// Serve Web UI
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Manual Detection API Endpoint
	http.HandleFunc("/api/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req mlclient.PredictRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Request JSON", http.StatusBadRequest)
			return
		}

		// Live prediction via Python ML Service
		resp, err := pythonClient.Predict(r.Context(), req)
		if err != nil {
			http.Error(w, "ML Engine Error: "+err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// OAuth Step 1: Redirect to Google
	http.HandleFunc("/auth/google", func(w http.ResponseWriter, r *http.Request) {
		url := gmailclient.GetLoginURL()
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	// OAuth Step 2: Google Callback Handler
	http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != gmailclient.OAuthStateString {
			http.Error(w, "Invalid State Token", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		emails, err := gmailclient.FetchRecentEmails(r.Context(), code)
		if err != nil {
			http.Error(w, "Failed to fetch Gmail data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Run predictions against Python server
		var results []GmailResult
		for _, email := range emails {
			resp, err := pythonClient.Predict(r.Context(), email)
			if err != nil {
				log.Printf("Failed to classify email %s: %v", email.MessageID, err)
				continue
			}

			results = append(results, GmailResult{
				MessageID:  email.MessageID,
				Subject:    email.Subject,
				Label:      resp.Label,
				Prediction: resp.Prediction,
			})
		}

		// Render HTML template with fetched results
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Results":   results,
			"ActiveTab": "gmail",
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	})

	log.Println("==================================================")
	log.Println(" BCA Minor Project Web Server Started")
	log.Println(" URL: http://127.0.0.1:8080")
	log.Println(" Integrated: Go Gateway -> Python ML Engine")
	log.Println("==================================================")

	log.Fatal(http.ListenAndServe(":8080", nil))
}