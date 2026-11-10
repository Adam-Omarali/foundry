package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"

	"foundry/backend/utils"

	"google.golang.org/api/idtoken"
)

// Google OAuth client IDs whose ID tokens we accept (the web app's and the
// Chrome extension's). An ID token's `aud` claim names the client it was
// issued to; checking it is what stops a token minted for some other app from
// being replayed against Foundry.
var googleClientIDs []string

var googleIssuers = []string{"accounts.google.com", "https://accounts.google.com"}

// InitGoogleAuth reads the accepted client IDs. Call once at startup.
//   - GOOGLE_CLIENT_ID: the web app's (also used by NextAuth). Required.
//   - GOOGLE_EXTENSION_CLIENT_ID: the Chrome extension's. Optional; without it
//     extension sign-ins are rejected.
func InitGoogleAuth() error {
	web := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	if web == "" {
		return errors.New("GOOGLE_CLIENT_ID environment variable is not set")
	}
	googleClientIDs = []string{web}
	if ext := strings.TrimSpace(os.Getenv("GOOGLE_EXTENSION_CLIENT_ID")); ext != "" {
		googleClientIDs = append(googleClientIDs, ext)
	}
	return nil
}

// VerifyGoogleIDToken checks a Google ID token's signature, expiry, issuer and
// audience, and returns the verified email address.
func VerifyGoogleIDToken(ctx context.Context, rawToken string) (string, error) {
	// Empty audience: we check it ourselves below against several client IDs.
	// idtoken.Validate checks signature + expiry, but NOT the issuer.
	payload, err := idtoken.Validate(ctx, rawToken, "")
	if err != nil {
		return "", fmt.Errorf("invalid id token: %w", err)
	}
	if !slices.Contains(googleIssuers, payload.Issuer) {
		return "", fmt.Errorf("unexpected issuer %q", payload.Issuer)
	}
	if !slices.Contains(googleClientIDs, payload.Audience) {
		return "", fmt.Errorf("unexpected audience %q", payload.Audience)
	}

	email, _ := payload.Claims["email"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)
	if email == "" || !verified {
		return "", errors.New("email missing or not verified")
	}
	return strings.ToLower(email), nil
}

// Authenticate extracts the Bearer token from the request and verifies it as a
// Foundry session JWT. It never touches the database.
func Authenticate(r *http.Request) (*utils.FoundryClaims, error) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		return nil, errors.New("missing bearer token")
	}
	return utils.VerifyJWT(token)
}
