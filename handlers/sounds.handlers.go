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

type SoundResponse struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Type        string `json:"type"` // 'background' or 'session'
	Tier        string `json:"tier"` // 'free' or 'premium'
	Duration    int    `json:"duration"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// GetSoundHandler godoc
// @Summary Get a single sound
// @Description Retrieves a sound by its UUID
// @Tags sounds
// @Accept json
// @Produce json
// @Param id path string true "UUID of the sound to retrieve"
// @Success 200 {object} models.Sound "Details of the sound"
// @Failure 404 {object} map[string]string "Sound not found"
// @Router /api/v1/sound/{id} [get]
func GetSoundHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]

	var sound models.Sound
	result := db.DB.Where("uuid = ?", userId).First(&sound)
	if result.Error != nil {
		http.Error(w, "Sound not found", http.StatusNotFound)
		return
	}

	// Convert SoundResponse
	response := SoundResponse{
		UUID:        sound.UUID,
		Name:        sound.Name,
		Type:        sound.Type,
		Tier:        sound.Tier,
		Duration:    sound.Duration,
		Description: sound.Description,
		URL:         sound.URL,
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetSoundsHandler godoc
// @Summary Get all sounds
// @Description Retrieves all sounds from the database
// @Tags sounds
// @Accept json
// @Produce json
// @Success 200 {array} SoundResponse "List of all sounds"
// @Router /api/v1/sound [get]
func GetSoundsHandler(w http.ResponseWriter, r *http.Request) {
	var sounds []models.Sound
	db.DB.Find(&sounds)

	// Convertir cada Sound a SoundResponse
	responses := make([]SoundResponse, len(sounds))
	for i, sound := range sounds {
		responses[i] = SoundResponse{
			UUID:        sound.UUID,
			Name:        sound.Name,
			Type:        sound.Type,
			Tier:        sound.Tier,
			Duration:    sound.Duration,
			Description: sound.Description,
			URL:         sound.URL,
		}
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(responses)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateSoundHandler godoc
// @Summary Create a sound
// @Description Adds a new sound to the database with detailed properties and an audio file.
// @Tags sounds
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Name of the sound"
// @Param type formData string true "Type of the sound; valid values are 'background' or 'session'."
// @Param tier formData string true "Tier of the sound; valid values are 'free' or 'premium'."
// @Param duration formData int true "Duration of the sound in seconds; must be a positive integer."
// @Param description formData string false "Description of the sound; optional but useful for additional context."
// @Param file formData file true "Audio file for the sound; required and must be in an acceptable audio format."
// @Success 201 {object} map[string]interface{} "Sound created successfully with details in the response body."
// @Failure 400 {object} map[string]string "Invalid input data with detailed error message."
// @Failure 409 {object} map[string]string "Name conflict; the specified sound name already exists."
// @Failure 500 {object} map[string]string "Internal server error; something went wrong on the server."
// @Router /api/v1/sound [post]
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

	// Valid types and tiers
	validTypes := map[string]bool{"background": true, "session": true}
	validTiers := map[string]bool{"free": true, "premium": true}

	// Validate sound type and tier
	if _, ok := validTypes[soundType]; !ok {
		http.Error(w, "Invalid type provided. Valid values are 'background' or 'session'.", http.StatusBadRequest)
		return
	}
	if _, ok := validTiers[tier]; !ok {
		http.Error(w, "Invalid tier provided. Valid values are 'free' or 'premium'.", http.StatusBadRequest)
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

// UpdateSoundHandler godoc
// @Summary Update a sound
// @Description Updates the properties of an existing sound identified by its UUID, except the sound file.
// @Tags sounds
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "UUID of the sound to update"
// @Param name formData string false "New name of the sound"
// @Param type formData string false "New type of the sound; valid values are 'background' or 'session'."
// @Param tier formData string false "New tier of the sound; valid values are 'free' or 'premium'."
// @Param duration formData int false "New duration of the sound in seconds; must be a positive integer if provided."
// @Param description formData string false "New description of the sound; optional and can provide additional context."
// @Success 200 {object} map[string]string "Sound updated successfully with message indicating successful update."
// @Failure 400 {object} map[string]string "Invalid input data with detailed error message."
// @Failure 404 {object} map[string]string "Sound not found if no sound matches the provided UUID."
// @Failure 500 {object} map[string]string "Internal server error if something goes wrong on the server side."
// @Router /api/v1/sound/{id} [put]
func UpdateSoundHandler(w http.ResponseWriter, r *http.Request) {

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

	// Valid types and tiers
	validTypes := map[string]bool{"background": true, "session": true}
	validTiers := map[string]bool{"free": true, "premium": true}

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

	// Validate sound type and tier if provided
	if soundType != "" && !validTypes[soundType] {
		http.Error(w, "Invalid type provided. Valid values are 'background' or 'session'.", http.StatusBadRequest)
		return
	}
	if tier != "" && !validTiers[tier] {
		http.Error(w, "Invalid tier provided. Valid values are 'free' or 'premium'.", http.StatusBadRequest)
		return
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

// DeleteSoundHandler godoc
// @Summary Delete a sound
// @Description Deletes a sound by UUID and removes the associated file from S3
// @Tags sounds
// @Accept json
// @Produce json
// @Param id path string true "UUID of the sound to delete"
// @Success 200 {object} map[string]string "Sound deleted successfully"
// @Failure 404 {object} map[string]string "Sound not found"
// @Failure 500 {object} map[string]string "Failed to delete sound or associated file"
// @Router /api/v1/sound/{id} [delete]
func DeleteSoundHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	var sound models.Sound
	result := db.DB.Where("uuid = ?", userId).First(&sound)
	if result.Error != nil {
		fmt.Println("Sound not found:", result.Error)
		http.Error(w, "Sound not found", http.StatusNotFound)
		return
	}

	// Extract bucket and key from URL
	bucket := bucketName // Replace with actual bucket name if needed
	key := sound.URL     // Assume URL is the full path if you store it as such

	// Delete the file from S3
	s3Client := services.NewS3Client()
	err := s3Client.DeleteFile(bucket, key)
	if err != nil {
		fmt.Println("Failed to delete audio from S3:", err)
		http.Error(w, "Failed to delete audio from S3", http.StatusInternalServerError)
		return
	}

	// Delete the sound from the database
	result = db.DB.Delete(&sound)
	if result.Error != nil {
		fmt.Println("Failed to delete sound:", result.Error)
		http.Error(w, "Failed to delete sound", http.StatusInternalServerError)
		return
	}

	fmt.Println("Sound deleted successfully:", sound)
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "Sound deleted successfully"})
	if err != nil {
		fmt.Println("Failed to write response:", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// HealthCheck godoc
// @Summary Check the health of the API
// @Description Returns the API status
// @Tags health
// @Accept  json
// @Produce  json
// @Success 200 {string} string "API is up and running"
// @Router /api/v1/health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("API is up and running"))
	if err != nil {
		fmt.Println(err)
		return
	}
}
