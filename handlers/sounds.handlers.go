package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"tbrBackend/db"
	"tbrBackend/helpers"
	"tbrBackend/models"
	"tbrBackend/services"
)

func GetSoundHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]

	var sound models.Sound
	result := db.DB.Where("uuid = ?", userId).First(&sound)
	if result.Error != nil {
		http.Error(w, "Sound not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(sound)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
	bucketName := os.Getenv("AWS_BUCKET_NAME")
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
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Failed to close file:", err)
		}
	}(file)

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
	url, err := s3Client.UploadFile(bucketName, sound.Type, handler.Filename, file)
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
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		fmt.Println("Failed to parse multipart form:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	userId := vars["id"]

	var sound models.Sound
	result := db.DB.Where("uuid = ?", userId).First(&sound)
	if result.Error != nil {
		fmt.Println("Sound not found:", result.Error)
		http.Error(w, "Sound not found", http.StatusNotFound)
		return
	}

	// Retrieve updated values from form
	name := r.FormValue("name")
	soundType := r.FormValue("type")
	tier := r.FormValue("tier")
	durationStr := r.FormValue("duration")
	description := r.FormValue("description")

	if name != "" {
		sound.Name = name
	}
	if soundType != "" {
		sound.Type = soundType
	}
	if tier != "" {
		sound.Tier = tier
	}
	if durationStr != "" {
		duration, err := strconv.Atoi(durationStr)
		if err != nil || duration <= 0 {
			fmt.Println("Invalid duration:", err)
			http.Error(w, "Invalid duration", http.StatusBadRequest)
			return
		}
		sound.Duration = duration
	}
	if description != "" {
		sound.Description = description
	}

	// Check for file update
	file, handler, err := r.FormFile("file")
	if err == nil {
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				fmt.Println("Failed to close file:", err)
			}
		}(file)
		// Upload the updated audio to S3
		s3Client := services.NewS3Client()
		url, err := s3Client.UploadFile(bucketName, sound.Type, handler.Filename, file)
		if err != nil {
			fmt.Println("Failed to upload audio to S3:", err)
			http.Error(w, "Failed to upload audio to S3", http.StatusInternalServerError)
			return
		}
		// Set the new URL returned by S3
		sound.URL = url
	}

	// Save updates to the database
	result = db.DB.Save(&sound)
	if result.Error != nil {
		fmt.Println("Failed to update sound:", result.Error)
		http.Error(w, "Failed to update sound", http.StatusInternalServerError)
		return
	}

	fmt.Println("Sound updated successfully:", sound)

	// Return a success response
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "Sound updated successfully"})
	if err != nil {
		fmt.Println("Failed to write response:", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
func DeleteSoundHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]

	var sound models.Sound
	result := db.DB.Where("uuid = ?", userId).First(&sound)
	if result.Error != nil {
		fmt.Println("Sound not found:", result.Error)
		http.Error(w, "Sound not found", http.StatusNotFound)
		return
	}

	result = db.DB.Delete(&sound)
	if result.Error != nil {
		fmt.Println("Failed to delete sound:", result.Error)
		http.Error(w, "Failed to delete sound", http.StatusInternalServerError)
		return
	}

	fmt.Println("Sound deleted successfully:", sound)

	// Return a success response
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string]string{"message": "Sound deleted successfully"})
	if err != nil {
		fmt.Println("Failed to write response:", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
