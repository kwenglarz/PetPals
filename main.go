package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"pet-management/models"
)

var pets = make([]models.Pet, 0)

var dataFile = "pets.json"

// load pets from the json file
func loadPets() {
	file, err := os.Open(dataFile)
	if err != nil {
		fmt.Println("No existing data found.")
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&pets)
	if err != nil {
		fmt.Println("Error loading data:", err)
	}
}

// save the pets to the json file
func savePets() {
	file, err := os.Create(dataFile)
	if err != nil {
		fmt.Println("Error saving data:", err)
		return
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(pets)
	if err != nil {
		fmt.Println("Error encoding data:", err)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// HOME HANDLER
func homeHandler(writer http.ResponseWriter, request *http.Request) {
	html, err:= template.ParseFiles("templates/home.html")
	check(err)
	err = html.Execute(writer, nil)
	check(err)
}

// ADD PET HANDLER
func addPetHandler(writer http.ResponseWriter, request *http.Request) {
    if request.Method == "POST" {
        // Get form data
        name := request.FormValue("name")
        petType := request.FormValue("type")
        vetName := request.FormValue("vet-name")
        vetAddress := request.FormValue("vet-address")
        vetPhone := request.FormValue("vet-phone")
        nextVetVisitStr := request.FormValue("next-vet-visit")
        vaccinations := request.FormValue("vaccinations")
        feedingFrequency := request.FormValue("feeding-frequency")
        foodType := request.FormValue("food-type")
        treatsQtyStr := request.FormValue("treats-qty")
        walkTime := request.FormValue("walk-time")
        playTime := request.FormValue("play-time")

        // Parse date
        nextVetVisit, err := time.Parse("2006-01-02", nextVetVisitStr)
        if err != nil {
            nextVetVisit = time.Time{} // Invalid date; set to zero time
        }

        // Convert treats quantity
        treatsQty, err := strconv.Atoi(treatsQtyStr)
        if err != nil {
            treatsQty = 0 // Default to 0 if invalid input
        }

        // Handle yes/no questions (checkboxes)
        cleanLitterbox := request.FormValue("clean-litterbox") == "y"
        nailTrim := request.FormValue("nail-trim") == "y"
        brushing := request.FormValue("brushing") == "y"
        haircut := request.FormValue("haircut") == "y"

        // Create the pet object
        newPet := models.Pet{
            Name:         name,
            Type:         petType,
            VetName:      vetName,
            VetAddress:   vetAddress,
            VetPhone:     vetPhone,
            NextVetVisit: nextVetVisit,
            Vaccinations: strings.Split(vaccinations, ","),
            Feeding: models.Feeding{
                Frequency: feedingFrequency,
                FoodType:  foodType,
                TreatsQty: treatsQty,
            },
            Routine: models.Routine{
                WalkTime: walkTime,
                PlayTime: playTime,
            },
            Maintenance: models.Maintenance{
                CleanLitterbox: models.MaintenanceItem{
                    Required: cleanLitterbox,
                    Frequency: request.FormValue("clean-litterbox-frequency"),
                },
                NailTrim: models.MaintenanceItem{
                    Required: nailTrim,
                    Frequency: request.FormValue("nail-trim-frequency"),
                },
                Brushing: models.MaintenanceItem{
                    Required: brushing,
                    Frequency: request.FormValue("brushing-frequency"),
                },
                Haircut: models.MaintenanceItem{
                    Required: haircut,
                    Frequency: request.FormValue("haircut-frequency"),
                },
            },
        }

        // Append the new pet to the pets list
        pets = append(pets, newPet)
		savePets()

        // Redirect or show success message
        http.Redirect(writer, request, "/view-pets", http.StatusFound)
    } else {
        // Display the Add Pet form

		html, err:= template.ParseFiles("templates/add-pet.html")
		check(err)
		err = html.Execute(writer, nil)
		check(err)
    }
}


// VIEW PETS HANDLER
func viewPetsHandler(writer http.ResponseWriter, request *http.Request) {
	// Create a PetsViewModel with the global `pets` slice
    model := models.PetsViewModel{Pets: pets}

	html, err:= template.ParseFiles("templates/view-pets.html")
	check(err)

	err = html.Execute(writer, model)
	check(err)
}

// UPDATE PET HANDLER
func updatePetsHandler(writer http.ResponseWriter, request *http.Request) {
	tmpl, err := template.ParseFiles("templates/update-pets.html")
	check(err)
	err = tmpl.Execute(writer, struct{ Pets []models.Pet }{Pets: pets})
	check(err)
}

// MODIFY PET HANDLER
func modifyPetHandler(w http.ResponseWriter, r *http.Request) {
	// Extract pet name from the URL
	petName := strings.TrimPrefix(r.URL.Path, "/modify-pet/")

	// Find the selected pet
	var selectedPet *models.Pet
	for i := range pets {
		if pets[i].Name == petName {
			selectedPet = &pets[i]
			break
		}
	}

	// If pet not found, return error
	if selectedPet == nil {
		http.Error(w, "Pet not found", http.StatusNotFound)
		return
	}

	// If it's a POST request, update the pet
	if r.Method == http.MethodPost {
		selectedPet.Name = r.FormValue("name")
		selectedPet.Type = r.FormValue("type")
		selectedPet.VetName = r.FormValue("vet-name")
		selectedPet.VetAddress = r.FormValue("vet-address")
		selectedPet.VetPhone = r.FormValue("vet-phone")

		VisitDateStr := r.FormValue("next-vet-visit")
		if VisitDateStr != "" {
			var err error
			selectedPet.NextVetVisit, err = time.Parse("2006-01-02", VisitDateStr)
			if err != nil {
				log.Fatal("Error parsing date: ", err)
			}
		}
		selectedPet.Feeding.Frequency = r.FormValue("feeding-frequency")
		selectedPet.Feeding.FoodType = r.FormValue("food-type")
		treatsQty, _ := strconv.Atoi(r.FormValue("treats-qty"))
		selectedPet.Feeding.TreatsQty = treatsQty

		selectedPet.Maintenance.CleanLitterbox.Required = r.FormValue("clean-litterbox") == "y"
		selectedPet.Maintenance.CleanLitterbox.Frequency = r.FormValue("clean-litterbox-frequency")
		selectedPet.Maintenance.NailTrim.Required = r.FormValue("nail-trim") == "y"
		selectedPet.Maintenance.NailTrim.Frequency = r.FormValue("nail-trim-frequency")
		selectedPet.Maintenance.Brushing.Required = r.FormValue("brushing") == "y"
		selectedPet.Maintenance.Brushing.Frequency = r.FormValue("brushing-frequency")
		selectedPet.Maintenance.Haircut.Required = r.FormValue("haircut") == "y"
		selectedPet.Maintenance.Haircut.Frequency = r.FormValue("haircut-frequency")

		selectedPet.Routine.WalkTime = r.FormValue("walk-time")
		selectedPet.Routine.PlayTime = r.FormValue("play-time")

		// Redirect after updating
		http.Redirect(w, r, "/view-pets", http.StatusFound)
	} else {
		// Load the update form with the selected pet data
		html, err := template.ParseFiles("templates/modify-pet.html")
		check(err)
		err = html.Execute(w, struct{ SelectedPet models.Pet }{SelectedPet: *selectedPet})
		check(err)
	}
}

// DELETE PETS HANDLER
func deletePetHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
        petNameToDelete := request.FormValue("pet-name")

		if petNameToDelete == "" {
			http.Error(writer, "Pet not found", http.StatusBadRequest)
		}

        deletePet(petNameToDelete)
		savePets()

        // After deleting, redirect to the same page to refresh the list
        http.Redirect(writer, request, "/delete-pet", http.StatusFound)
    } else {
		// Create a PetsViewModel with the global `pets` slice
		model := models.PetsViewModel{Pets: pets}

		html, err:= template.ParseFiles("templates/delete-pet.html")
		check(err)
		err = html.Execute(writer, model)
		check(err)
	}
}

// DELETE PET BY NAME
func deletePet(petName string) {
    for i, pet := range pets {
        if pet.Name == petName {
            pets = append(pets[:i], pets[i+1:]...)  // Remove the pet from the slice
            fmt.Println("Deleted pet:", petName)
            break
        }
    }
}

func main() {	
	loadPets()
 
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/view-pets", viewPetsHandler)
	http.HandleFunc("/add-pet", addPetHandler)
	http.HandleFunc("/delete-pet", deletePetHandler)
	http.HandleFunc("/update-pets", updatePetsHandler)
	http.HandleFunc("/modify-pet/", modifyPetHandler)
	
	fmt.Println("Starting server on :8080...")
	fmt.Println("Navigate to localhost:8080/home to open Pet Pals!")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}


// selectedPet.Maintenance.CleanLitterbox = r.FormValue("clean-litter-box")