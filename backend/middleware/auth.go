package middleware

import (
	"encoding/json"
	"fmt"
	"foundry/backend/db"
	"net/http"
	"os"
)

func VerifyGoogleToken(token string) (map[string]interface{}, error) {
	res, err := http.Get("https://oauth2.googleapis.com/tokeninfo?access_token=" + token)
	if err != nil || res.StatusCode != 200 {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer res.Body.Close()

	var info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	fmt.Println("Info:", info)
	// ✅ Check audience matches your Chrome app's client ID
	expectedAudience := os.Getenv("GOOGLE_CLIENT_ID")
	if expectedAudience == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID environment variable not set")
	}
	if info["aud"] != expectedAudience {
		return nil, http.ErrNoCookie
	}

	db.InsetUser(info["email"].(string), "free")

	return info, nil
}
