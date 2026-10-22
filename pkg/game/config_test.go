package game

import (
	"strings"
	"testing"
)

func TestValidateActiveAccounts(t *testing.T) {
	first := AccountConfig{ID: "one", Credentials: GameCredentials{Email: "player@example.invalid", Token: "secret-placeholder"}}
	for _, tc := range []struct {
		name   string
		second AccountConfig
		valid  bool
	}{
		{"distinct", AccountConfig{ID: "two"}, true},
		{"duplicate ID", AccountConfig{ID: "one"}, false},
		{"blank ID", AccountConfig{}, false},
		{"whitespace ID", AccountConfig{ID: " two "}, false},
		{"duplicate login", AccountConfig{ID: "two", Credentials: GameCredentials{Email: " PLAYER@EXAMPLE.INVALID "}}, false},
		{"duplicate token", AccountConfig{ID: "two", Credentials: GameCredentials{Token: "secret-placeholder"}}, false},
		{"invalid bind", AccountConfig{ID: "two", BindIP: "invalid"}, false},
		{"valid bind", AccountConfig{ID: "two", BindIP: "127.0.0.1"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateActiveAccounts([]AccountConfig{first, tc.second})
			if (err == nil) != tc.valid {
				t.Fatalf("erro=%v válido=%v", err, tc.valid)
			}
			if err != nil && (strings.Contains(err.Error(), first.Credentials.Token) || strings.Contains(err.Error(), first.Credentials.Email)) {
				t.Fatal("erro expôs credencial")
			}
		})
	}
}
