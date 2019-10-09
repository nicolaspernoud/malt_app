package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nicolaspernoud/malt-app/internal/auth"
	"github.com/nicolaspernoud/malt-app/internal/models"

	"github.com/alexedwards/scs/v2"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/qor/admin"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/database"
	"github.com/qor/roles"
)

var sessionManager *scs.SessionManager

func main() {
	// Start the server
	fmt.Println("Listening on: http://localhost:8081/admin?locale=fr-FR")
	http.ListenAndServe(":8081", createMux())
}

func createMux() http.Handler {
	// Init the session manager and pass it to the auth package
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	auth.Init()
	auth.InitSM(sessionManager)

	// Gather the models than will be managed by QOR
	models := models.Export()

	// Set up the business database
	DB, _ := gorm.Open("sqlite3", "./data/business.db")
	DB.AutoMigrate(models...)

	// Initialize
	Admin := admin.New(&admin.AdminConfig{
		DB:       DB,
		SiteName: "Malt App",
		Auth:     &auth.Auth{AuthLoginURL: "/OAuth2Login", AuthLogoutURL: "/logout"},
	})

	// Create resources from GORM-backend model
	for _, s := range models {
		Admin.AddResource(s, &admin.Config{
			Permission: roles.Allow(roles.Read, roles.Anyone).Allow(roles.CRUD, "admin"),
		})
	}

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

	// Delete models object (enable it to be garbage collected)
	models = nil

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
