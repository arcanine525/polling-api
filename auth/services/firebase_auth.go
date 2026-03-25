package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// firebaseSignInRequest is the request body for Firebase Auth REST API.
type firebaseSignInRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

// FirebaseSignInResponse holds the response from Firebase Auth REST API.
type FirebaseSignInResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"`
	Email        string `json:"email"`
}

// firebaseErrorResponse holds Firebase REST API error details.
type firebaseErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// SignInWithEmailPassword authenticates a user via Firebase Auth REST API.
func SignInWithEmailPassword(apiKey, email, password string) (*FirebaseSignInResponse, error) {
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", apiKey)

	body, err := json.Marshal(firebaseSignInRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to call Firebase Auth: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var fbErr firebaseErrorResponse
		if json.Unmarshal(respBody, &fbErr) == nil && fbErr.Error.Message != "" {
			return nil, fmt.Errorf("%s", mapFirebaseError(fbErr.Error.Message))
		}
		return nil, fmt.Errorf("firebase auth failed with status %d", resp.StatusCode)
	}

	var result FirebaseSignInResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// mapFirebaseError maps Firebase error codes to user-friendly messages.
func mapFirebaseError(code string) string {
	switch code {
	case "EMAIL_NOT_FOUND":
		return "invalid email or password"
	case "INVALID_PASSWORD":
		return "invalid email or password"
	case "USER_DISABLED":
		return "account has been disabled"
	case "INVALID_LOGIN_CREDENTIALS":
		return "invalid email or password"
	case "TOO_MANY_ATTEMPTS_TRY_LATER":
		return "too many failed attempts, please try again later"
	default:
		return "authentication failed"
	}
}
