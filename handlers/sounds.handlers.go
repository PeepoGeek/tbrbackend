package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"tbrBackend/db"
	"tbrBackend/helpers"
	"tbrBackend/models"
	"tbrBackend/services"
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

func CreateSoundHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		fmt.Println("Failed to parse multipart form:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	soundType := r.FormValue("type")
	tier := r.FormValue("tier")
	durationStr := r.FormValue("duration")
	description := r.FormValue("description")
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Failed to retrieve file:", err)
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate the input data
	missingFields := []string{}
	if name == "" {
		missingFields = append(missingFields, "name")
	}
	if soundType == "" {
		missingFields = append(missingFields, "type")
	}
	if tier == "" {
		missingFields = append(missingFields, "tier")
	}
	if durationStr == "" {
		missingFields = append(missingFields, "duration")
	}
	if len(missingFields) > 0 {
		message := "Missing required fields: " + helpers.StringJoin(missingFields, ", ")
		fmt.Println(message)
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		fmt.Println("Invalid duration:", err)
		http.Error(w, "Invalid duration", http.StatusBadRequest)
		return
	}

	sound := models.Sound{
		UUID:        uuid.New().String(),
		Name:        name,
		Type:        soundType,
		Tier:        tier,
		Duration:    duration,
		Description: description,
	}

	// Check for unique constraint violation
	var existingSound models.Sound
	result := db.DB.Where("name = ?", sound.Name).First(&existingSound)
	if result.Error == nil {
		fmt.Println("Failed to create sound: name already exists")
		http.Error(w, "Name already exists", http.StatusConflict)
		return
	}

	// Upload the audio to S3
	s3Client := services.NewS3Client()
	url, err := s3Client.UploadFile("tbrsoundbucket", sound.Type, handler.Filename, file)
	if err != nil {
		fmt.Println("Failed to upload audio to S3:", err)
		http.Error(w, "Failed to upload audio to S3", http.StatusInternalServerError)
		return
	}

	// Set the URL returned by S3
	sound.URL = url

	// Create the sound in the database
	fmt.Println("Creating sound:", sound)
	result = db.DB.Create(&sound)
	if result.Error != nil {
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
