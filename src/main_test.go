package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"strings"
	"time"
	"fmt"
	"os"
	"pet-management/models"
)

var testPet = models.Pet{
	Name:       "Buddy",
	Type:       "Dog",
	VetName:    "Dr. Smith",
	VetAddress: "123 Vet St.",
	VetPhone:   "555-1234",
	Feeding: models.Feeding{
		Frequency: "Twice a day",
		FoodType:  "Dry",
		TreatsQty: 2,
	},
	NextVetVisit: time.Now().AddDate(0, 1, 0), // One month from now
	Routine: models.Routine{
		WalkTime: "8:00 AM",
		PlayTime: "6:00 PM",
	},
}

var testFormDataForPet = url.Values{
	"name":                 {"Buddy"},
	"type":                 {"Dog"},
	"vet-name":             {"Dr. Smith"},
	"vet-address":          {"123 Pet St"},
	"vet-phone":            {"123-456-7890"},
	"next-vet-visit":       {"2024-12-01"},
	"vaccinations":         {"Rabies,Parvo"},
	"feeding-frequency":    {"Twice a day"},
	"food-type":            {"Dry food"},
	"treats-qty":           {"2"},
	"walk-time":            {"Morning"},
	"play-time":            {"Evening"},
	"clean-litterbox":      {"y"},
	"clean-litterbox-frequency": {"Daily"},
	"nail-trim":            {"y"},
	"nail-trim-frequency":  {"Weekly"},
	"brushing":             {"y"},
	"brushing-frequency":   {"Weekly"},
	"haircut":              {"y"},
	"haircut-frequency":    {"Monthly"},
}

func TestMain(m *testing.M) {
	fmt.Println("Setup before tests")

	// Set data file to test file
	dataFile = "pets-test.json"

	code := m.Run()

	emptyFile(dataFile)

	os.Exit(code)
}

// Function cleans file used for testing
func emptyFile(filePath string) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic("Failed to open file: " + err.Error())
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		panic("Failed to truncate file: " + err.Error())
	}
}

// Function reset the pets array, used to clean up state
func resetPets() {
    pets = []models.Pet{}
}

func TestHomeHandler(t *testing.T) {
	resetPets() //isolate each test
	// Create a request to pass to the handler
	request, err := http.NewRequest("GET", "/home", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a response recorder to capture the handler's response
	writer := httptest.NewRecorder()

	// Call the handler
	homeHandler(writer, request)

	response := writer.Result()
	// Check the status code
	if response.StatusCode != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", response.StatusCode, http.StatusOK)
	}

	// // Check the response body
	expectedContent := "Welcome to Pet Pals!" // Example content from home.html
	if !strings.Contains(writer.Body.String(), expectedContent) {
		t.Errorf("Handler returned unexpected body: got %v want content to include %v", writer.Body.String(), expectedContent)
	}
}


// testing view pets handler with no pets
func TestViewPetsHandler_NoPets(t *testing.T) {
	resetPets() //isolate each test
	// Ensure the pets slice is empty
	pets = []models.Pet{}

	// Create a request to pass to the handler
	request, err := http.NewRequest("GET", "/view-pets", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a response recorder to capture the handler's response
	recorder := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(viewPetsHandler)
	handler.ServeHTTP(recorder, request)

	// Check the status code
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body for the "no pets" message
	expectedContent := "No pets available to display."
	if !strings.Contains(recorder.Body.String(), expectedContent) {
		t.Errorf("Handler returned unexpected body: got %v want content to include %v", recorder.Body.String(), expectedContent)
	}
}

func TestViewPetsHandler_WithPets(t *testing.T) {
	resetPets() //isolate each test
	// Set up sample pets for the test
	pets = []models.Pet{
		testPet,
	}

	// Create a request to pass to the handler
	request, err := http.NewRequest("GET", "/view-pets", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a response recorder to capture the handler's response
	recorder := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(viewPetsHandler)
	handler.ServeHTTP(recorder, request)

	// Check the status code
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body for some expected content
	expectedContent := "Buddy" // Verify that the pet's name appears in the rendered HTML
	if !strings.Contains(recorder.Body.String(), expectedContent) {
		t.Errorf("Handler returned unexpected body: got %v want content to include %v", recorder.Body.String(), expectedContent)
	}
}

func TestAddPetHandler(t *testing.T) {
	resetPets() //isolate each test

	// Create a POST request
	// Create the pet and ensure it was made
	request, err := http.NewRequest("POST", "/add-pet", strings.NewReader(testFormDataForPet.Encode()))
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Record the response
	recorder := httptest.NewRecorder()
	http.HandlerFunc(addPetHandler).ServeHTTP(recorder, request)
	// Assert the response status code
	if recorder.Code != http.StatusFound {
		t.Errorf("Expected status code 302, got %d", recorder.Code)
	}
	if len(pets) != 1 {
		t.Fatalf("Expected 1 pet in the pets slice, got %d", len(pets))
	}

	// Validate the added pet's data
	addedPet := pets[0]
	if addedPet.Name != "Buddy" || addedPet.Type != "Dog" {
		t.Errorf("Unexpected pet data: %+v", addedPet)
	}
	if len(addedPet.Vaccinations) != 2 || addedPet.Vaccinations[0] != "Rabies" {
		t.Errorf("Vaccinations not parsed correctly: %+v", addedPet.Vaccinations)
	}
	if addedPet.Feeding.Frequency != "Twice a day" || addedPet.Feeding.TreatsQty != 2 {
		t.Errorf("Feeding details incorrect: %+v", addedPet.Feeding)
	}
	if !addedPet.Maintenance.CleanLitterbox.Required || addedPet.Maintenance.CleanLitterbox.Frequency != "Daily" {
		t.Errorf("Maintenance details incorrect: %+v", addedPet.Maintenance)
	}
	if addedPet.Routine.WalkTime != "Morning" || addedPet.Routine.PlayTime != "Evening" {
		t.Errorf("Routine details incorrect: %+v", addedPet.Routine)
	}
}

func TestDeletePetHandler(t *testing.T) {
    // Setup a sample pet list for testing
    resetPets()
	pets = []models.Pet{
		testPet,
	}

	// View delete pet page
	request, err := http.NewRequest("POST", "/delete-pet", strings.NewReader(testFormDataForPet.Encode()))
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// Create a response recorder to capture the handler's response
	recorder := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(deletePetHandler)
	handler.ServeHTTP(recorder, request)

	// Check the status code
	if status := recorder.Code; status != http.StatusFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
}

func TestDeletePetViewWithPetsHandler(t *testing.T) {
    // Setup a sample pet list for testing
    resetPets()

	// Set up sample pets for the test
	pets = []models.Pet{
		testPet,
	}

	// View delete pet page
	request, err := http.NewRequest("GET", "/delete-pet", nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}
	
	// Create a response recorder to capture the handler's response
	recorder := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(deletePetHandler)
	handler.ServeHTTP(recorder, request)

	// Check the status code
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body for the "no pets" message
	expectedContent := "Buddy"
	if !strings.Contains(recorder.Body.String(), expectedContent) {
		t.Errorf("Handler returned unexpected body: got %v want content to include %v", recorder.Body.String(), expectedContent)
	}
}

func TestDeletePetViewWithoutPetsHandler(t *testing.T) {
    // Setup a sample pet list for testing
    resetPets()

	// View delete pet page
	request, err := http.NewRequest("GET", "/delete-pet", nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}
	
	// Create a response recorder to capture the handler's response
	recorder := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(deletePetHandler)
	handler.ServeHTTP(recorder, request)

	// Check the status code
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body for the "no pets" message
	expectedContent := "No pets available to display."
	if !strings.Contains(recorder.Body.String(), expectedContent) {
		t.Errorf("Handler returned unexpected body: got %v want content to include %v", recorder.Body.String(), expectedContent)
	}
}


func TestModifyPetHandler(t *testing.T) {
	// Setup initial data
	resetPets()
	pets = []models.Pet{
		testPet,
	}

	// Create updated form data to update Buddy
	updatedBuddyData := bytes.NewBufferString(
		"name=Buddy&type=Dog&vet-name=Dr. Bark&vet-address=456 Dog St.&vet-phone=555-9876&feeding-frequency=Once a day&food-type=Canned food&treats-qty=3&walk-time=Afternoon&play-time=Night",
	)

	// Make request to update Buddy
	request, err := http.NewRequest("POST", "/modify-pet/Buddy", updatedBuddyData)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a ResponseRecorder to capture the response
	recorder := httptest.NewRecorder()

	// Call the ModifyPetHandler
	handler := http.HandlerFunc(modifyPetHandler)
	handler.ServeHTTP(recorder, request)

	// Check for redirection (i.e., successful modification)
	if status := recorder.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}

	// Check that the pet's details were updated
	// The pet with name "Buddy" should have been updated
	var updatedPet *models.Pet
	for i := range pets {
		if pets[i].Name == "Buddy" {
			updatedPet = &pets[i]
			break
		}
	}

	// Ensure the pet exists and the fields were updated
	if updatedPet == nil {
		t.Fatal("Pet not found after modification")
	}

	// Check updated fields
	if updatedPet.Type != "Dog" {
		t.Errorf("expected Type to be 'Dog', got '%s'", updatedPet.Type)
	}
	if updatedPet.VetName != "Dr. Bark" {
		t.Errorf("expected VetName to be 'Dr. Bark', got '%s'", updatedPet.VetName)
	}
	if updatedPet.VetAddress != "456 Dog St." {
		t.Errorf("expected VetAddress to be '456 Dog St.', got '%s'", updatedPet.VetAddress)
	}
	if updatedPet.VetPhone != "555-9876" {
		t.Errorf("expected VetPhone to be '555-9876', got '%s'", updatedPet.VetPhone)
	}
	if updatedPet.Feeding.Frequency != "Once a day" {
		t.Errorf("expected Feeding Frequency to be 'Once a day', got '%s'", updatedPet.Feeding.Frequency)
	}
	if updatedPet.Feeding.FoodType != "Canned food" {
		t.Errorf("expected FoodType to be 'Canned food', got '%s'", updatedPet.Feeding.FoodType)
	}
	if updatedPet.Feeding.TreatsQty != 3 {
		t.Errorf("expected TreatsQty to be '3', got '%d'", updatedPet.Feeding.TreatsQty)
	}
	if updatedPet.Routine.WalkTime != "Afternoon" {
		t.Errorf("expected WalkTime to be 'Afternoon', got '%s'", updatedPet.Routine.WalkTime)
	}
	if updatedPet.Routine.PlayTime != "Night" {
		t.Errorf("expected PlayTime to be 'Night', got '%s'", updatedPet.Routine.PlayTime)
	}
}

func TestModifyPetViewHandler(t *testing.T) {
    // Setup a sample pet list for testing
    resetPets()

	// Set up sample pets for the test
	pets = []models.Pet{
		testPet,
	}

	// View modify pet page
	request, err := http.NewRequest("GET", "/modify-pet/Buddy", nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}
	
	// Create a response recorder to capture the handler's response
	recorder := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(deletePetHandler)
	handler.ServeHTTP(recorder, request)

	// Check the status code
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body for the "no pets" message
	expectedContent := "Buddy"
	if !strings.Contains(recorder.Body.String(), expectedContent) {
		t.Errorf("Handler returned unexpected body: got %v want content to include %v", recorder.Body.String(), expectedContent)
	}
}

func TestUpdatePetViewWithPetsHandler(t *testing.T) {
    // Setup a sample pet list for testing
    resetPets()

	// Set up sample pets for the test
	pets = []models.Pet{
		testPet,
	}

	// View delete pet page
	request, err := http.NewRequest("GET", "/update-pets", nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}
	
	// Create a response recorder to capture the handler's response
	recorder := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(deletePetHandler)
	handler.ServeHTTP(recorder, request)

	// Check the status code
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body for the "no pets" message
	expectedContent := "Buddy"
	if !strings.Contains(recorder.Body.String(), expectedContent) {
		t.Errorf("Handler returned unexpected body: got %v want content to include %v", recorder.Body.String(), expectedContent)
	}
}

func TestUpdatePetViewWithoutPetsHandler(t *testing.T) {
    // Setup a sample pet list for testing
    resetPets()

	// View delete pet page
	request, err := http.NewRequest("GET", "/update-pets", nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}
	
	// Create a response recorder to capture the handler's response
	recorder := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(deletePetHandler)
	handler.ServeHTTP(recorder, request)

	// Check the status code
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body for the "no pets" message
	expectedContent := "No pets available to display."
	if !strings.Contains(recorder.Body.String(), expectedContent) {
		t.Errorf("Handler returned unexpected body: got %v want content to include %v", recorder.Body.String(), expectedContent)
	}
}
