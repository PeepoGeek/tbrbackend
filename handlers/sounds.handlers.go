package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"tbrBackend/db"
	"tbrBackend/helpers"
	"tbrBackend/models"
)

func GetSoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("GetSounds"))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func GetSoundsHandler(w http.ResponseWriter, r *http.Request) {
	var sound []models.Sound
	db.DB.Find(&sound)
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(sound)
	if err != nil {
		fmt.Println(err)
		return
	}

}

// CreateSoundHandler handles the creation of a new sound.
// CreateSoundHandler handles the creation of a new sound.
func CreateSoundHandler(w http.ResponseWriter, r *http.Request) {
	var sound models.Sound
	err := json.NewDecoder(r.Body).Decode(&sound)
	if err != nil {
		fmt.Println("Failed to decode request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate the input data
	missingFields := []string{}
	if sound.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if sound.Type == "" {
		missingFields = append(missingFields, "type")
	}
	if sound.Tier == "" {
		missingFields = append(missingFields, "tier")
	}
	if sound.Duration <= 0 {
		missingFields = append(missingFields, "duration")
	}
	if sound.URL == "" {
		missingFields = append(missingFields, "url")
	}

	if len(missingFields) > 0 {
		message := "Missing required fields: " + helpers.StringJoin(missingFields, ", ")
		fmt.Println(message)
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	sound.UUID = uuid.New().String()

	// Create the sound in the database
	fmt.Println("Creating sound:", sound)
	result := db.DB.Create(&sound)

	// Check for unique constraint violation
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			fmt.Println("Failed to create sound: name already exists")
			http.Error(w, "Name already exists", http.StatusConflict)
			return
		}
		fmt.Println("Failed to create sound:", result.Error)
		http.Error(w, "Failed to create sound", http.StatusInternalServerError)
		return
	}

	fmt.Println("Sound created successfully:", sound)

	// Return a success response
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "Sound created successfully"})
	if err != nil {
		fmt.Println("Failed to write response:", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func UpdateSoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("UpdateSound"))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func DeleteSoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("DeleteSound"))
	if err != nil {
		fmt.Println(err)
		return
	}
}
