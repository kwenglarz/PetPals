package models

import (
	"time"
)

type Pet struct {
	Type         string
	Name         string
	VetName      string
	VetAddress   string
	VetPhone     string
	NextVetVisit time.Time
	Vaccinations []string
	Feeding      Feeding
	Maintenance  Maintenance
	Routine      Routine
}

type PetsViewModel struct {
    Pets []Pet
}

type Feeding struct {
	Frequency string
	FoodType  string
	TreatsQty int
}

type MaintenanceItem struct {
	Required  bool
	Frequency string
}
type Maintenance struct {
	CleanLitterbox MaintenanceItem
	NailTrim       MaintenanceItem
	Brushing       MaintenanceItem
	Haircut        MaintenanceItem
}

type Routine struct {
	WalkTime string
	PlayTime string
}