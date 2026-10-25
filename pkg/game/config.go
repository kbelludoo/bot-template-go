package game

import (
	"fmt"
	"net"
	"strings"
)

// ValidateActiveAccounts rejeita sessões duplicadas antes de iniciar a frota.
// Deve receber apenas as contas selecionadas para execução.
func ValidateActiveAccounts(accounts []AccountConfig) error {
	ids, emails, tokens := make(map[string]bool), make(map[string]bool), make(map[string]bool)
	for index, account := range accounts {
		id := strings.TrimSpace(account.ID)
		if id == "" || id != account.ID || ids[id] {
			return fmt.Errorf("conta #%d: ID vazio, duplicado ou com espaços nas extremidades", index+1)
		}
		ids[id] = true
		if account.BindIP != "" && net.ParseIP(account.BindIP) == nil {
			return fmt.Errorf("conta #%d: bindIP inválido", index+1)
		}
		email := strings.ToLower(strings.TrimSpace(account.Credentials.Email))
		if email != "" {
			if emails[email] {
				return fmt.Errorf("conta #%d: credencial de login já utilizada por outra conta ativa", index+1)
			}
			emails[email] = true
		}
		token := strings.TrimSpace(account.Credentials.Token)
		if token != "" {
			if tokens[token] {
				return fmt.Errorf("conta #%d: token já utilizado por outra conta ativa", index+1)
			}
			tokens[token] = true
		}
	}
	return nil
}
