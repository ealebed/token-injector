package gcp

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestIDToken_GetDuration(t *testing.T) {
	token := IDToken{}

	tests := []struct {
		name      string
		jwtToken  string
		wantErr   bool
		checkFunc func(time.Duration) bool
	}{
		{
			name:     "valid token with future expiration",
			jwtToken: createTestJWT(t, time.Now().Add(1*time.Hour)),
			wantErr:  false,
			checkFunc: func(d time.Duration) bool {
				// Should be approximately 1 hour, allow 5 minute tolerance
				return d > 55*time.Minute && d < 65*time.Minute
			},
		},
		{
			name:     "valid token with 30 minutes expiration",
			jwtToken: createTestJWT(t, time.Now().Add(30*time.Minute)),
			wantErr:  false,
			checkFunc: func(d time.Duration) bool {
				// Should be approximately 30 minutes, allow 2 minute tolerance
				return d > 28*time.Minute && d < 32*time.Minute
			},
		},
		{
			name:     "valid token with past expiration",
			jwtToken: createTestJWT(t, time.Now().Add(-1*time.Hour)),
			wantErr:  false,
			checkFunc: func(d time.Duration) bool {
				// Should be negative (expired)
				return d < 0
			},
		},
		{
			name:     "invalid JWT format",
			jwtToken: "invalid.jwt.token",
			wantErr:  true,
		},
		{
			name:     "empty token",
			jwtToken: "",
			wantErr:  true,
		},
		{
			name:     "malformed JWT - missing parts",
			jwtToken: "header.payload",
			wantErr:  true,
		},
		{
			name:     "JWT without exp claim",
			jwtToken: createTestJWTWithoutExp(t),
			wantErr:  true,
		},
		{
			name:     "JWT with invalid exp claim type",
			jwtToken: createTestJWTWithInvalidExp(t),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := token.GetDuration(tt.jwtToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("IDToken.GetDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				if !tt.checkFunc(got) {
					t.Errorf("IDToken.GetDuration() = %v, checkFunc returned false", got)
				}
			}
		})
	}
}

// createTestJWT creates a valid JWT token with the specified expiration time
func createTestJWT(t *testing.T, exp time.Time) string {
	t.Helper()

	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	claims := jwt.MapClaims{
		"exp": json.Number(strconv.FormatInt(exp.Unix(), 10)),
		"iat": json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		"sub": "test@example.com",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("failed to marshal header: %v", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal claims: %v", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create unsigned JWT (we don't need a valid signature for testing)
	return headerB64 + "." + claimsB64 + "."
}

// createTestJWTWithoutExp creates a JWT token without exp claim
func createTestJWTWithoutExp(t *testing.T) string {
	t.Helper()

	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	claims := jwt.MapClaims{
		"iat": json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		"sub": "test@example.com",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("failed to marshal header: %v", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal claims: %v", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	return headerB64 + "." + claimsB64 + "."
}

// createTestJWTWithInvalidExp creates a JWT token with invalid exp claim (not a number)
func createTestJWTWithInvalidExp(t *testing.T) string {
	t.Helper()

	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	claims := jwt.MapClaims{
		"exp": "not-a-number",
		"iat": json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		"sub": "test@example.com",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("failed to marshal header: %v", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal claims: %v", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	return headerB64 + "." + claimsB64 + "."
}
