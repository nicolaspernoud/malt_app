package models

import "github.com/jinzhu/gorm"

// Employee has and belongs to many languages, use `employee_languages` as join table
type Employee struct {
	gorm.Model
	Name          string
	Languages     []*Language `gorm:"many2many:employee_languages;"`
	HomeAddresses []HomeAddress
}

// Language and belongs to many languages, use `employee_languages` as join table
type Language struct {
	gorm.Model
	Name      string
	Employees []*Employee `gorm:"many2many:employee_languages;"`
}

// HomeAddress as a many to one relationship with Employee
type HomeAddress struct {
	gorm.Model
	Name       string
	City       string
	EmployeeID uint
}

// Export generates an array of the models we want in the admin interface
// Add here the models that you want QOR admin to manage
func Export() []interface{} {
	return []interface{}{&Employee{}, &Language{}, &HomeAddress{}}
}
