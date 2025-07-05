package middleware

import (
	"encoding/json"
	"fmt"
	"foundry/backend/db"
	"io"
	"net/http"
)

func VerifyGoogleToken(token string) (map[string]interface{}, error) {
	res, err := http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + token)
	if err != nil {
		fmt.Println("Error making request to Google:", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		fmt.Printf("Error response from Google: %s\n", string(body))
		return nil, fmt.Errorf("invalid token: status code %d", res.StatusCode)
	}

	var info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		fmt.Println("Error decoding response:", err)
		return nil, err
	}

	// Verify the email is verified
	if verified, ok := info["email_verified"].(bool); !ok || !verified {
		fmt.Println("Email not verified")
		return nil, fmt.Errorf("email not verified")
	}

	// Get the email
	email, ok := info["email"].(string)
	if !ok {
		fmt.Println("Email not found in user info")
		return nil, fmt.Errorf("email not found in token")
	}

	// Insert new user or update existing user's last sign-in
	userId := db.GetUserId(email)
	if userId == -1 {
		db.InsertUser(email, "free")
	}
	db.SignInUser(email)

	return info, nil
}
