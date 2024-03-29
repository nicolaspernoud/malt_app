// This is a mock oauth2 server for development purposes

package mockoauth2

import (
	"fmt"
	"net/http"
)

// CreateMockOAuth2 creates a mock OAuth2 serve mux for development purposes
func CreateMockOAuth2() *http.ServeMux {

	mux := http.NewServeMux()

	// Returns authorization code back to the user
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		redir := query.Get("redirect_uri") + "?state=" + query.Get("state") + "&code=mock_code"
		http.Redirect(w, r, redir, 302)
	})

	// Returns access token back to the user
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write([]byte("access_token=mocktoken&scope=user&token_type=bearer"))
	})

	// Returns userinfo back to the user
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"displayName": "Us ER",
			"memberOf": [
				"CN=ADMIN",
				"CN=OTHERS"
			],
			"id": "aLongId==",
			"login": "USER"
		}`))
	})

	// Logout
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Logout OK")
	})

	return mux
}
