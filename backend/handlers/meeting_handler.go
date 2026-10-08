package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"meet-ai/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type MeetingHandler struct {
	DB *pgx.Conn
}

type TranscriptionResponse struct {
	Filename            string  `json:"filename"`
	Language            string  `json:"language"`
	LanguageProbability float64 `json:"language_probability"`
	Transcript          string  `json:"transcript"`
}

func NewMeetingHandler(db *pgx.Conn) *MeetingHandler {
	return &MeetingHandler{
		DB: db,
	}
}

func (h *MeetingHandler) CreateMeeting(c *gin.Context) {
	var meeting models.Meeting

	if err := c.ShouldBindJSON(&meeting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	query := `
		INSERT INTO meetings (
			title,
			description,
			host_name,
			host_email,
			scheduled_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, host_name, host_email,
		          scheduled_at, status, audio_path, transcript, created_at
	`

	err := h.DB.QueryRow(
		context.Background(),
		query,
		meeting.Title,
		meeting.Description,
		meeting.HostName,
		meeting.HostEmail,
		meeting.ScheduledAt,
	).Scan(
		&meeting.ID,
		&meeting.Title,
		&meeting.Description,
		&meeting.HostName,
		&meeting.HostEmail,
		&meeting.ScheduledAt,
		&meeting.Status,
		&meeting.AudioPath,
		&meeting.Transcript,
		&meeting.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create meeting",
		})
		return
	}

	c.JSON(http.StatusCreated, meeting)
}

func (h *MeetingHandler) GetMeetings(c *gin.Context) {
	query := `
		SELECT
			id,
			title,
			description,
			host_name,
			host_email,
			scheduled_at,
			status,
			audio_path,
			created_at
		FROM meetings
		ORDER BY scheduled_at DESC
	`

	rows, err := h.DB.Query(context.Background(), query)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch meetings",
		})
		return
	}

	defer rows.Close()

	meetings := []models.Meeting{}

	for rows.Next() {
		var meeting models.Meeting

		err := rows.Scan(
			&meeting.ID,
			&meeting.Title,
			&meeting.Description,
			&meeting.HostName,
			&meeting.HostEmail,
			&meeting.ScheduledAt,
			&meeting.Status,
			&meeting.AudioPath,
			&meeting.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read meeting data",
			})
			return
		}

		meetings = append(meetings, meeting)
	}

	c.JSON(http.StatusOK, meetings)
}

func (h *MeetingHandler) UploadMeetingAudio(c *gin.Context) {

	meetingID := c.Param("id")

	id, err := strconv.Atoi(meetingID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid meeting ID",
		})
		return
	}

	// Check whether meeting exists
	var existingID int

	err = h.DB.QueryRow(
		context.Background(),
		"SELECT id FROM meetings WHERE id = $1",
		id,
	).Scan(&existingID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Meeting not found",
		})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("audio")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Audio file is required",
		})
		return
	}

	// Create uploads directory if it doesn't exist
	err = os.MkdirAll("uploads", os.ModePerm)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create uploads directory",
		})
		return
	}

	// Generate unique filename
	extension := filepath.Ext(file.Filename)

	filename := fmt.Sprintf(
		"meeting_%d_%d%s",
		id,
		time.Now().Unix(),
		extension,
	)

	filePath := filepath.Join("uploads", filename)

	// Save file
	err = c.SaveUploadedFile(file, filePath)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save audio file",
		})
		return
	}

	// Update database
	query := `
		UPDATE meetings
		SET audio_path = $1
		WHERE id = $2
	`

	_, err = h.DB.Exec(
		context.Background(),
		query,
		filePath,
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update meeting",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Meeting audio uploaded successfully",
		"meeting_id": id,
		"file_name":  file.Filename,
		"audio_path": filePath,
	})
}

func (h *MeetingHandler) TranscribeMeeting(c *gin.Context) {

	meetingID := c.Param("id")

	id, err := strconv.Atoi(meetingID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid meeting ID",
		})
		return
	}

	// Get audio path from database
	var audioPath string

	err = h.DB.QueryRow(
		context.Background(),
		"SELECT audio_path FROM meetings WHERE id = $1",
		id,
	).Scan(&audioPath)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Meeting not found",
		})
		return
	}

	if audioPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No audio file uploaded for this meeting",
		})
		return
	}

	// Open audio file
	audioFile, err := os.Open(audioPath)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open audio file",
		})
		return
	}

	defer audioFile.Close()

	// Create multipart request
	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)

	fileName := filepath.Base(audioPath)

	part, err := writer.CreateFormFile("file", fileName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to prepare audio file",
		})
		return
	}

	_, err = io.Copy(part, audioFile)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read audio file",
		})
		return
	}

	err = writer.Close()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to prepare transcription request",
		})
		return
	}

	// Send audio to Python AI service
	response, err := http.Post(
		"http://127.0.0.1:8000/transcribe",
		writer.FormDataContentType(),
		&requestBody,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AI transcription service unavailable",
		})
		return
	}

	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read AI service response",
		})
		return
	}

	if response.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Transcription failed",
			"details": string(responseData),
		})
		return
	}

	var transcription TranscriptionResponse

	err = json.Unmarshal(responseData, &transcription)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid response from AI service",
		})
		return
	}

	// Save transcript to database
	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET transcript = $1 WHERE id = $2",
		transcription.Transcript,
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save transcript",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Meeting transcribed successfully",
		"meeting_id": id,
		"language":   transcription.Language,
		"transcript": transcription.Transcript,
	})
}
