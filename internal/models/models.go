package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/nicolaspernoud/malt-app/internal/auth"
	"github.com/qor/admin"
	"github.com/qor/roles"
)

var (
	// DB is the application business database
	DB *gorm.DB
)

// Recipe is a beer recipe
type Recipe struct {
	gorm.Model
	Name    string
	Batches []Batch
}

// Batch is made from a recipe and has a list of events altering (or not) it's volume
type Batch struct {
	gorm.Model
	RecipeID      uint
	Step          string
	StartVolume   int
	CurrentVolume int
	Events        []Event
	Transfers     []Transfer
	VolumeShares  string
}

// Event is attached to a batch and can alter its volume
type Event struct {
	gorm.Model
	BatchID uint
	Name    string
	Volume  int
}

// Transfer is a special event attached to a batch and can alter its volume shares
type Transfer struct {
	gorm.Model
	BatchID uint
	To      Container `gorm:"foreignkey:ToID"`
	ToID    uint
	From    Container `gorm:"foreignkey:FromID"`
	FromID  uint
	Volume  int
}

// Container is where the beer is stored
type Container struct {
	gorm.Model
	Name   string
	Volume int
}

// AfterFind calculates the batch volumes according to the events
func (b *Batch) AfterFind() (err error) {
	// Work out the fermentation tank volume
	var events []Event
	DB.Where("batch_id = ?", b.ID).Find(&events)
	b.CurrentVolume = b.StartVolume
	for _, e := range events {
		b.CurrentVolume += e.Volume
	}
	// Work out the shares between containers ; TODO : check that volumes are less than the fermentation tank volume
	// Work out what the volumes will be (array of structs{container, volume})
	var containers []Container
	DB.Find(&containers)
	m := make(map[string]int)
	var transfers []Transfer
	DB.Preload("From").Preload("To").Where("batch_id = ?", b.ID).Find(&transfers)
	// For each volumes sustract the froms and add the to
	for _, t := range transfers {
		m[t.From.Name] -= t.Volume
		m[t.To.Name] += t.Volume
	}
	var shares string
	for key, val := range m {
		shares += fmt.Sprintf("%s: %s, ", key, strconv.Itoa(val))
	}
	b.VolumeShares = strings.TrimSuffix(shares, ", ")
	return
}

// CreateAdmin creates an admin based on the models
func CreateAdmin(siteName string) *admin.Admin {
	// Set up the business database
	DB, _ = gorm.Open("sqlite3", "./data/business.db")
	models := []interface{}{&Recipe{}, &Batch{}, &Event{}, &Transfer{}, &Container{}}
	DB.AutoMigrate(models...)

	// Initialize
	Admin := admin.New(&admin.AdminConfig{
		DB:       DB,
		SiteName: siteName,
		Auth:     &auth.Auth{AuthLoginURL: "/OAuth2Login", AuthLogoutURL: "/logout"},
	})

	// Create resources from GORM-backend model
	for _, s := range models {
		Admin.AddResource(s, &admin.Config{
			Permission: roles.Allow(roles.Read, roles.Anyone).Allow(roles.CRUD, "admin"),
		})
	}

	batch := Admin.GetResource("Batch")
	batch.Meta(&admin.Meta{Name: "Step", Type: "select_one", Config: &admin.SelectOneConfig{Collection: []string{"mixed", "brewed", "fermented"}}})

	return Admin
}
