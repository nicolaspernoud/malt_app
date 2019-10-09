package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/nicolaspernoud/malt-app/internal/auth"
	"github.com/nicolaspernoud/malt-app/tester"
)

func setupMockOAuthServer() *httptest.Server {
	mux := http.NewServeMux()

	// Returns authorization code back to the user, but without the provided state
	mux.HandleFunc("/auth-wrong-state", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		redir := query.Get("redirect_uri") + "?state=" + "a-random-state" + "&code=mock_code"
		http.Redirect(w, r, redir, 302)
	})

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
				"CN=GGD_ORG_DG-DEES-DINSI_TOUS,OU=ORGA,OU=APPLICATIONS,DC=ben,DC=oscar,DC=gly",
				"CN=GGD_ORG_DG-DEES-DINSI-DAAG_TOUS,OU=ORGA,OU=APPLICATIONS,DC=ben,DC=oscar,DC=gly"
			],
			"id": "aLongId==",
			"login": "USER"
		}`))
	})

	// Returns userinfo back to the user (with an admin user)
	mux.HandleFunc("/admininfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"displayName": "Ad MIN",
			"memberOf": [
				"CN=GGD_ORG_DG-DEES-DINSI_TOUS,OU=ORGA,OU=APPLICATIONS,DC=ben,DC=oscar,DC=gly",
				"CN=GGD_PASI_ADMIN_GROUP,OU=ORGA,OU=APPLICATIONS,DC=ben,DC=oscar,DC=gly"
			],
			"id": "anotherLongId==",
			"login": "ADMIN"
		}`))
	})

	// Logout
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Logout OK")
	})

	server := httptest.NewServer(mux)
	return server
}

func TestEndToEnd(t *testing.T) {

	// Create the mock OAuth2 server
	oAuth2Server := setupMockOAuthServer()
	defer oAuth2Server.Close()

	// Create the server
	mainMux := createMux()
	ts := httptest.NewServer(mainMux)
	defer ts.Close()
	url, _ := url.Parse(ts.URL)
	port := url.Port()

	// Set the constants with environment variables
	os.Setenv("CLIENT_ID", "clientid")
	os.Setenv("CLIENT_SECRET", "clientsecret")
	os.Setenv("AUTH_URL", oAuth2Server.URL+"/auth")
	os.Setenv("TOKEN_URL", oAuth2Server.URL+"/token")
	os.Setenv("USERINFO_URL", oAuth2Server.URL+"/userinfo")
	os.Setenv("LOGOUT_URL", oAuth2Server.URL+"/logout")
	os.Setenv("REDIRECT_URL", "http://localhost:"+port+"/OAuth2Callback")
	os.Setenv("ADMIN_GROUP", "GGD_PASI_ADMIN_GROUP")

	// Create the cookie jars
	userJar, _ := cookiejar.New(nil)

	// Set the constants

	// Security tests (this tests are to check that the security protections works)
	// Set the server to access failing OAuth2 endpoints
	os.Setenv("AUTH_URL", oAuth2Server.URL+"/auth-wrong-state")
	auth.Init()
	// Test that if the OAuth2server doesn't return the correct state, the login fails
	tester.DoRequestOnServer(t, userJar, port, "GET", "/admin", "", "", 500, "invalid oauth state")

	// Unlogged tests (those tests chexks the behaviour for an unconnected user)
	// Set the server to access normal OAuth2 endpoints
	os.Setenv("AUTH_URL", oAuth2Server.URL+"/auth")
	auth.Init()
	// Try the healthcheck (must pass)
	tester.DoRequestOnServer(t, userJar, port, "GET", "/healthcheck", "", "", 200, "OK")

	// Normal users tests (those tests checks the normal behaviour for an user)
	// Try to login (must pass)
	tester.DoRequestOnServer(t, userJar, port, "GET", "/admin", "", "", 200, "")
	// Try to access something (must pass)
	tester.DoRequestOnServer(t, userJar, port, "GET", "/admin/employees.json", "", "", 200, `[]`)

	// Logout
	tester.DoRequestOnServer(t, userJar, port, "GET", "/logout", "", "Logout OK", 200, "")
}
