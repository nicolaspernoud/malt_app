package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/nicolaspernoud/malt_app/internal/auth"
	"github.com/nicolaspernoud/malt_app/internal/mockoauth2"
	"github.com/nicolaspernoud/malt_app/internal/models"

	"github.com/alexedwards/scs/v2"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/qor/admin"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/database"
)

var (
	sessionManager *scs.SessionManager
	debugMode      = flag.Bool("debug", false, "Debug mode, enables mock OAuth2 server")
)

func main() {
	// Parse flags
	flag.Parse()
	// Start a mock oauth2 server if debug mode is on
	if *debugMode {
		mockOAuth2Port := ":8090"
		go http.ListenAndServe(mockOAuth2Port, mockoauth2.CreateMockOAuth2())
		fmt.Println("Mock OAuth2 server Listening on: http://localhost" + mockOAuth2Port)
	}

	// Start the server
	httpPort := ":8081"
	fmt.Println("Listening on: http://localhost" + httpPort + "/admin?locale=fr-FR")
	http.ListenAndServe(httpPort, createMux())
}

func createMux() http.Handler {
	// Init the session manager and pass it to the auth package
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	auth.Init()
	auth.InitSM(sessionManager)

	// Attach models to Admin
	Admin := models.CreateAdmin("Malt App")

	// Set up translations
	i18ndb, _ := gorm.Open("sqlite3", "./data/i18n.db")
	I18n := i18n.New(
		database.New(i18ndb), // load translations from the database,
	)
	Admin.AddResource(I18n, &admin.Config{Menu: []string{"Settings"}})

	// Initalize an HTTP request multiplexer
	mux := http.NewServeMux()

	// Mount admin to the mux
	Admin.MountTo("/admin", mux)

	// Handle all other routes with the multiplexer
	mux.HandleFunc("/OAuth2Login", auth.HandleOAuth2Login)
	mux.HandleFunc("/OAuth2Callback", auth.HandleOAuth2Callback)
	mux.HandleFunc("/logout", auth.Logout)
	mux.HandleFunc("/api/userinfo", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			auth.SendUser(w, req)
			return
		}
		http.Error(w, "method not allowed", 405)
	})
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "OK")
	})

	return (sessionManager.LoadAndSave(mux))
}
