package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
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
	StartVolume   int
	CurrentVolume int
	Events        []Event
}

// AfterCreate alters the batch volume according to the created event
func (b *Batch) AfterCreate(tx *gorm.DB) (err error) {
	tx.Model(b).Update("current_volume", b.StartVolume)
	return
}

// Event is attached to a batch and can alter its volume
type Event struct {
	gorm.Model
	BatchID uint
	Name    string
	Volume  int
}

// AfterCreate alters the batch volume according to the created event
func (e *Event) AfterCreate(tx *gorm.DB) (err error) {
	var b Batch
	tx.Model(&Batch{}).Where("ID = ?", e.BatchID).First(&b)
	b.CurrentVolume += e.Volume
	tx.Save(&b)
	return
}

// Export generates an array of the models we want in the admin interface
// Add here the models that you want QOR admin to manage
func Export() []interface{} {
	return []interface{}{&Recipe{}, &Batch{}, &Event{}}
}

// CustomizeAdmin customize the admin fields
func CustomizeAdmin(Admin *admin.Admin) {
	/*recipe := Admin.GetResource("Recipe")
	recipe.Meta(&admin.Meta{Name: "Batches", Type: "select_many", Config: &admin.SelectManyConfig{SelectMode: "bottom_sheet"}})

	batch := Admin.GetResource("Batch")
	batch.Meta(&admin.Meta{Name: "Events", Type: "select_many", Config: &admin.SelectManyConfig{SelectMode: "bottom_sheet"}})
	//batch.Meta(&admin.Meta{Name: "RecipeID", Type: "select_one"})*/

	//event := Admin.GetResource("Event")
	//event.Meta(&admin.Meta{Name: "BatchID", Type: "select_one"})
}
