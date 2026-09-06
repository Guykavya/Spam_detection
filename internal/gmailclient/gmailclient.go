package gmailclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"spam-email-detector/go-backend/karan/internal/mlclient"
)

// OAuth2 Configuration setup
var OAuthConfig = &oauth2.Config{
	ClientID:     "354292213188-irvighttkpmgqkgi7d743c6h5pm0qc0m.apps.googleusercontent.com", // Replace with Cloud Console Credentials
	ClientSecret: "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", // Replace with Cloud Console Credentials
	RedirectURL:  "http://127.0.0.1:8080/auth/google/callback",
	Scopes:       []string{gmail.GmailReadonlyScope},
	Endpoint:     google.Endpoint,
}

// State token for CSRF protection
const OAuthStateString = "random-bca-project-state-token"

// GetLoginURL generates the Google OAuth URL
func GetLoginURL() string {
	return OAuthConfig.AuthCodeURL(OAuthStateString, oauth2.AccessTypeOffline)
}

// EmailMessage contains the fetched email metadata and body content
type EmailMessage struct {
	MessageID string `json:"message_id"`
	Subject   string `json:"subject"`
	Snippet   string `json:"snippet"`
	Body      string `json:"body"`
}

// FetchRecentEmails obtains the authorization token and reads recent emails
func FetchRecentEmails(ctx context.Context, code string) ([]mlclient.PredictRequest, error) {
	token, err := OAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %v", err)
	}

	client := OAuthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}

	// Fetch up to 5 emails from Inbox
	user := "me"
	res, err := srv.Users.Messages.List(user).MaxResults(5).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	var emailRequests []mlclient.PredictRequest

	for _, m := range res.Messages {
		msg, err := srv.Users.Messages.Get(user, m.Id).Format("full").Do()
		if err != nil {
			continue
		}

		subject, body := extractEmailData(msg)
		emailRequests = append(emailRequests, mlclient.PredictRequest{
			MessageID: msg.Id,
			Subject:   subject,
			Body:      body,
		})
	}

	return emailRequests, nil
}

// Extract Subject header and parse MIME payload body
func extractEmailData(msg *gmail.Message) (string, string) {
	subject := "No Subject"
	if msg.Payload != nil {
		for _, header := range msg.Payload.Headers {
			if strings.EqualFold(header.Name, "Subject") {
				subject = header.Value
				break
			}
		}
	}

	body := parseMessagePart(msg.Payload)
	if body == "" {
		body = msg.Snippet
	}

	return subject, body
}

// Recursively decode URL-Safe Base64 text from MIME parts
func parseMessagePart(part *gmail.MessagePart) string {
	if part == nil {
		return ""
	}

	if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
		// Use RawURLEncoding to handle Gmail's unpadded base64 strings
		decoded, err := base64.RawURLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return string(decoded)
		}
		// Fallback to standard URLEncoding
		decoded, err = base64.URLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return string(decoded)
		}
	}

	for _, subPart := range part.Parts {
		text := parseMessagePart(subPart)
		if text != "" {
			return text
		}
	}

	return ""
}