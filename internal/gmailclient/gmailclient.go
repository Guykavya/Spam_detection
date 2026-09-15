package gmailclient

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"spam-email-detector/go-backend/karan/internal/mlclient"
)

const (
	oauthStateCookieName = "spam_detector_oauth_state"
	oauthStateMaxAge     = 10 * 60
)

// oauthConfig loads credentials at runtime so no OAuth secret is committed.
func oauthConfig() (*oauth2.Config, error) {
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	redirectURL := configuredRedirectURL()

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, fmt.Errorf(
			"Google OAuth client credentials and redirect URL must be configured",
		)
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{gmail.GmailReadonlyScope},
		Endpoint:     google.Endpoint,
	}, nil
}

func configuredRedirectURL() string {
	if redirectURL := strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL")); redirectURL != "" {
		return redirectURL
	}

	if renderHostname := strings.TrimSpace(os.Getenv("RENDER_EXTERNAL_HOSTNAME")); renderHostname != "" {
		return "https://" + renderHostname + "/auth/google/callback"
	}

	return ""
}

func useSecureCookies() bool {
	return strings.HasPrefix(
		strings.ToLower(configuredRedirectURL()),
		"https://",
	)
}

// GetLoginURL generates a fresh OAuth state token and stores it in a
// short-lived, HTTP-only cookie for verification after Google's redirect.
func GetLoginURL(w http.ResponseWriter) (string, error) {
	config, err := oauthConfig()
	if err != nil {
		return "", err
	}

	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate OAuth state: %w", err)
	}

	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   oauthStateMaxAge,
		HttpOnly: true,
		Secure:   useSecureCookies(),
		SameSite: http.SameSiteLaxMode,
	})

	return config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// ValidateOAuthState protects the callback against cross-site request forgery.
func ValidateOAuthState(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return false
	}

	returnedState := r.URL.Query().Get("state")
	valid := returnedState != "" && subtle.ConstantTimeCompare(
		[]byte(returnedState),
		[]byte(cookie.Value),
	) == 1

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   useSecureCookies(),
		SameSite: http.SameSiteLaxMode,
	})

	return valid
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
	config, err := oauthConfig()
	if err != nil {
		return nil, err
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %v", err)
	}

	client := config.Client(ctx, token)
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
