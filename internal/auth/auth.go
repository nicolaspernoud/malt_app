package auth

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nicolaspernoud/ninicobox-v3-server/pkg/common"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"golang.org/x/oauth2"

	"github.com/alexedwards/scs/v2"
	"github.com/qor/roles"
)

var (
	oauth2Config   *oauth2.Config
	sessionManager *scs.SessionManager
)

// Init initialize the configuration
func Init() {
	oauth2Config = &oauth2.Config{
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"login", "memberOf", "displayName", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  os.Getenv("AUTH_URL"),
			TokenURL: os.Getenv("TOKEN_URL"),
		},
	}

	// Register the custom type with the encoding/gob package
	gob.Register(User{})

	// Register the function to work out if an user has the admin role
	roles.Register("admin", func(r *http.Request, currentUser interface{}) bool {
		user, err := GetUser(r)
		return err == nil && user.Login != "" && user.IsAdmin
	})
}

// InitSM init the current package with the provided SCS session manager
func InitSM(sm *scs.SessionManager) {
	sessionManager = sm
}

// User struct implements the QOR  qor.CurrentUser interface https://godoc.org/github.com/qor/qor#CurrentUser
type User struct {
	Login    string   `json:"login"`
	FullName string   `json:"displayName,omitempty"`
	MemberOf []string `json:"memberOf"`
	IsAdmin  bool     `json:"isAdmin"`
	Name     string   `json:"name,omitempty"`
	Email    string   `json:"email,omitempty"`
}

// DisplayName returns the user full name
func (u User) DisplayName() string {
	if u.FullName != "" {
		return u.FullName
	}
	return u.Name
}

// Auth implements the qor Auth interface, and can be used in an admin (https://doc.getqor.com/admin/authentication.html, https://godoc.org/github.com/qor/admin#Auth)
type Auth struct {
	AuthLoginURL  string
	AuthLogoutURL string
}

// GetCurrentUser returns the current user from the context
func (a Auth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	user, err := GetUser(c.Request)
	if err != nil || user.Login == "" {
		return nil
	}
	return user
}

// LoginURL gets the login url for the admin to redirect on auth error (no user)
func (a Auth) LoginURL(c *admin.Context) string {
	return a.AuthLoginURL
}

// LogoutURL sets the logout url that is inserted into the admin page
func (a Auth) LogoutURL(c *admin.Context) string {
	return a.AuthLogoutURL
}

// HandleOAuth2Login handles the OAuth2 login
func HandleOAuth2Login(w http.ResponseWriter, r *http.Request) {
	// Generate state and store it in session
	oauthStateString, err := common.GenerateRandomString(48)
	if err != nil {
		log.Fatalf("Error generating OAuth2 strate string :%v\n", err)
	}
	sessionManager.Put(r.Context(), "oauthStateString", oauthStateString)
	url := oauth2Config.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleOAuth2Callback handles the OAuth2 Callback and get user info
func HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	// Recover state from session
	oauthStateString := sessionManager.GetString(r.Context(), "oauthStateString")

	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Error(w, "invalid oauth state", http.StatusInternalServerError)
		return
	}

	code := r.FormValue("code")
	token, err := oauth2Config.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("Code exchange failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", os.Getenv("USERINFO_URL")+"?access_token="+token.AccessToken, nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	response, err := client.Do(req)
	if err != nil || response.StatusCode == http.StatusBadRequest {
		fmt.Printf("User info failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var user User
	if response.Body == nil {
		http.Error(w, "no response body", 400)
		return
	}
	err = json.NewDecoder(response.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	for key, group := range user.MemberOf {
		user.MemberOf[key] = strings.TrimPrefix(strings.Split(group, ",")[0], "CN=")
	}

	sessionManager.Put(r.Context(), "user", user)

	http.Redirect(w, r, "/admin?locale=fr-FR", http.StatusFound)
}

// Logout remove the user from the cookie store
func Logout(w http.ResponseWriter, r *http.Request) {
	// Delete session
	sessionManager.Remove(r.Context(), "user")
	http.Redirect(w, r, os.Getenv("LOGOUT_URL"), http.StatusTemporaryRedirect)
}

// SendUser returns the user found from the request into the store
func SendUser(w http.ResponseWriter, r *http.Request) {
	user, err := GetUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(user)
}

// IsMemberOf checks if an user is a member of the given group
func (u User) IsMemberOf(group string) bool {
	for _, ugroup := range u.MemberOf {
		if ugroup == group {
			return true
		}
	}
	return false
}

// GetUser gets an user from a request
func GetUser(r *http.Request) (User, error) {
	user, ok := sessionManager.Get(r.Context(), "user").(User)
	if !ok {
		return User{}, errors.New("type assertion to 'User' failed")
	}
	user.IsAdmin = user.IsMemberOf(os.Getenv("ADMIN_GROUP"))
	return user, nil
}
