// Package basicauth provides ...
package basicauth

import (
	"fmt"
	"net/http"
	"strings"

	auth "github.com/abbot/go-http-auth"
)

var (
	realm               = "hirproxy"
	authorizationHeader = "Authorization"
	headerField         = "X-Hirproxy-User"
)

type middleware struct {
	next  http.Handler
	users map[string]string
	auth  *auth.BasicAuth

	ignoreCredential bool
}

// New - create new basic auth
func New(next http.Handler, rawUsers string, ignoreCredential bool) (http.Handler, error) {
	userMap, err := getUsers(rawUsers)
	if err != nil {
		return nil, err
	}
	m := &middleware{
		next:             next,
		users:            userMap,
		ignoreCredential: ignoreCredential,
	}
	m.auth = auth.NewBasicAuthenticator(realm, m.secretBasic)
	return m, nil
}

func (m *middleware) secretBasic(user, realm string) string {
	if secret, ok := m.users[user]; ok {
		return secret
	}
	return ""
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.auth.Wrap(func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		if m.ignoreCredential {
			r.Header.Del(authorizationHeader)
		}
		r.Header[headerField] = []string{r.Username}
		m.next.ServeHTTP(w, &r.Request)
	}).ServeHTTP(w, r)
}

func getUsers(rawUsers string) (map[string]string, error) {
	users := getLinesFromRaw(rawUsers)
	userMap := make(map[string]string)
	for _, user := range users {
		username, userHash, err := basicParser(user)
		if err != nil {
			return userMap, err
		}
		userMap[username] = userHash
	}
	return userMap, nil
}

func basicParser(user string) (string, string, error) {
	split := strings.Split(user, ":")
	if len(split) != 2 {
		return "", "", fmt.Errorf("error parsing BasicUser: %v", user)
	}
	return split[0], split[1], nil
}

func getLinesFromRaw(rawUsers string) []string {
	var filteredLines []string
	for _, rawLine := range strings.Split(rawUsers, "\n") {
		line := strings.TrimSpace(rawLine)
		if line != "" && !strings.HasPrefix(line, "#") {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines
}
