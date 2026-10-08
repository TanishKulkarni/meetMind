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

type AnalysisResponse struct {
	Status   string                 `json:"status"`
	Analysis map[string]interface{} `json:"analysis"`
}

func NewMeetingHandler(db *pgx.Conn) *MeetingHandler {
	return &MeetingHandler{
		DB: db,
	}
}

// ============================================================
// HELPER: TRANSCRIBE AUDIO
// ============================================================

func (h *MeetingHandler) transcribeAudio(audioPath string) (TranscriptionResponse, error) {

	audioFile, err := os.Open(audioPath)

	if err != nil {
		return TranscriptionResponse{}, err
	}

	defer audioFile.Close()

	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)

	fileName := filepath.Base(audioPath)

	part, err := writer.CreateFormFile("file", fileName)

	if err != nil {
		return TranscriptionResponse{}, err
	}

	_, err = io.Copy(part, audioFile)

	if err != nil {
		return TranscriptionResponse{}, err
	}

	err = writer.Close()

	if err != nil {
		return TranscriptionResponse{}, err
	}

	response, err := http.Post(
		"http://127.0.0.1:8000/transcribe",
		writer.FormDataContentType(),
		&requestBody,
	)

	if err != nil {
		return TranscriptionResponse{}, err
	}

	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		return TranscriptionResponse{}, err
	}

	if response.StatusCode != http.StatusOK {
		return TranscriptionResponse{}, fmt.Errorf(
			"transcription service returned status %d: %s",
			response.StatusCode,
			string(responseData),
		)
	}

	var transcription TranscriptionResponse

	err = json.Unmarshal(responseData, &transcription)

	if err != nil {
		return TranscriptionResponse{}, err
	}

	return transcription, nil
}

// ============================================================
// HELPER: ANALYZE TRANSCRIPT
// ============================================================

func (h *MeetingHandler) analyzeTranscript(
	transcript string,
) (map[string]interface{}, error) {

	requestBodyData := map[string]string{
		"transcript": transcript,
	}

	requestBody, err := json.Marshal(requestBodyData)

	if err != nil {
		return nil, err
	}

	response, err := http.Post(
		"http://127.0.0.1:8000/analyze",
		"application/json",
		bytes.NewBuffer(requestBody),
	)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"analysis service returned status %d: %s",
			response.StatusCode,
			string(responseData),
		)
	}

	var analysisResponse AnalysisResponse

	err = json.Unmarshal(responseData, &analysisResponse)

	if err != nil {
		return nil, err
	}

	return analysisResponse.Analysis, nil
}

// ============================================================
// CREATE MEETING
// ============================================================

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
		RETURNING
			id,
			title,
			description,
			host_name,
			host_email,
			scheduled_at,
			status,
			audio_path,
			transcript,
			created_at
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

// ============================================================
// GET ALL MEETINGS
// ============================================================

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
			transcript,
			analysis,
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
			&meeting.Transcript,
			&meeting.Analysis,
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

func (h *MeetingHandler) GetMeeting(c *gin.Context) {
	meetingID := c.Param("id")

	id, err := strconv.Atoi(meetingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid meeting ID",
		})
		return
	}

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
			transcript,
			analysis,
			created_at
		FROM meetings
		WHERE id = $1
	`

	var meeting models.Meeting

	err = h.DB.QueryRow(
		context.Background(),
		query,
		id,
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
		&meeting.Analysis,
		&meeting.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Meeting not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch meeting",
		})
		return
	}

	c.JSON(http.StatusOK, meeting)
}

// ============================================================
// UPLOAD MEETING AUDIO
// ============================================================

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

		if err == pgx.ErrNoRows {

			c.JSON(http.StatusNotFound, gin.H{
				"error": "Meeting not found",
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check meeting",
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

	// Create uploads directory

	err = os.MkdirAll(
		"uploads",
		os.ModePerm,
	)

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

	filePath := filepath.Join(
		"uploads",
		filename,
	)

	// Save file

	err = c.SaveUploadedFile(
		file,
		filePath,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save audio file",
		})

		return
	}

	// Update database

	query := `
		UPDATE meetings
		SET
			audio_path = $1,
			status = $2
		WHERE id = $3
	`

	_, err = h.DB.Exec(
		context.Background(),
		query,
		filePath,
		"uploaded",
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

// ============================================================
// TRANSCRIBE MEETING
// ============================================================

func (h *MeetingHandler) TranscribeMeeting(c *gin.Context) {

	meetingID := c.Param("id")

	id, err := strconv.Atoi(meetingID)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid meeting ID",
		})

		return
	}

	// Get audio path

	var audioPath string

	err = h.DB.QueryRow(
		context.Background(),
		"SELECT audio_path FROM meetings WHERE id = $1",
		id,
	).Scan(&audioPath)

	if err != nil {

		if err == pgx.ErrNoRows {

			c.JSON(http.StatusNotFound, gin.H{
				"error": "Meeting not found",
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch meeting",
		})

		return
	}

	if audioPath == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No audio file uploaded for this meeting",
		})

		return
	}

	// Update status

	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET status = $1 WHERE id = $2",
		"transcribing",
		id,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update meeting status",
		})

		return
	}

	// Transcribe

	transcription, err := h.transcribeAudio(
		audioPath,
	)

	if err != nil {

		h.DB.Exec(
			context.Background(),
			"UPDATE meetings SET status = $1 WHERE id = $2",
			"failed",
			id,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Transcription failed",
			"details": err.Error(),
		})

		return
	}

	// Save transcript

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

	// Update status

	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET status = $1 WHERE id = $2",
		"transcribed",
		id,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update meeting status",
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

// ============================================================
// ANALYZE MEETING
// ============================================================

func (h *MeetingHandler) AnalyzeMeeting(c *gin.Context) {

	meetingID := c.Param("id")

	id, err := strconv.Atoi(meetingID)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid meeting ID",
		})

		return
	}

	// Get transcript

	var transcript string

	err = h.DB.QueryRow(
		context.Background(),
		"SELECT transcript FROM meetings WHERE id = $1",
		id,
	).Scan(&transcript)

	if err != nil {

		if err == pgx.ErrNoRows {

			c.JSON(http.StatusNotFound, gin.H{
				"error": "Meeting not found",
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch transcript",
		})

		return
	}

	if transcript == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Meeting has not been transcribed yet",
		})

		return
	}

	// Update status

	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET status = $1 WHERE id = $2",
		"analyzing",
		id,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update meeting status",
		})

		return
	}

	// Analyze transcript

	analysis, err := h.analyzeTranscript(
		transcript,
	)

	if err != nil {

		h.DB.Exec(
			context.Background(),
			"UPDATE meetings SET status = $1 WHERE id = $2",
			"failed",
			id,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Meeting analysis failed",
			"details": err.Error(),
		})

		return
	}

	// Convert analysis to JSON

	analysisJSON, err := json.Marshal(
		analysis,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process analysis",
		})

		return
	}

	// Save analysis

	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET analysis = $1 WHERE id = $2",
		analysisJSON,
		id,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save meeting analysis",
		})

		return
	}

	// Completed

	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET status = $1 WHERE id = $2",
		"completed",
		id,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update meeting status",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Meeting analyzed successfully",
		"meeting_id": id,
		"analysis":   analysis,
	})
}

// ============================================================
// PROCESS COMPLETE MEETING
// ============================================================

func (h *MeetingHandler) ProcessMeeting(c *gin.Context) {

	meetingID := c.Param("id")

	id, err := strconv.Atoi(meetingID)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid meeting ID",
		})

		return
	}

	// Get audio path

	var audioPath string

	err = h.DB.QueryRow(
		context.Background(),
		"SELECT audio_path FROM meetings WHERE id = $1",
		id,
	).Scan(&audioPath)

	if err != nil {

		if err == pgx.ErrNoRows {

			c.JSON(http.StatusNotFound, gin.H{
				"error": "Meeting not found",
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch meeting",
		})

		return
	}

	if audioPath == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No audio file uploaded for this meeting",
		})

		return
	}

	// ========================================================
	// STEP 1: TRANSCRIBING
	// ========================================================

	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET status = $1 WHERE id = $2",
		"transcribing",
		id,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update meeting status",
		})

		return
	}

	// ========================================================
	// STEP 2: TRANSCRIBE
	// ========================================================

	transcription, err := h.transcribeAudio(
		audioPath,
	)

	if err != nil {

		h.DB.Exec(
			context.Background(),
			"UPDATE meetings SET status = $1 WHERE id = $2",
			"failed",
			id,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Transcription failed",
			"details": err.Error(),
		})

		return
	}

	// Save transcript

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

	// ========================================================
	// STEP 3: ANALYZING
	// ========================================================

	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET status = $1 WHERE id = $2",
		"analyzing",
		id,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update meeting status",
		})

		return
	}

	// ========================================================
	// STEP 4: ANALYZE
	// ========================================================

	analysis, err := h.analyzeTranscript(
		transcription.Transcript,
	)

	if err != nil {

		h.DB.Exec(
			context.Background(),
			"UPDATE meetings SET status = $1 WHERE id = $2",
			"failed",
			id,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Meeting analysis failed",
			"details": err.Error(),
		})

		return
	}

	// Convert analysis to JSON

	analysisJSON, err := json.Marshal(
		analysis,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process analysis",
		})

		return
	}

	// Save analysis

	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET analysis = $1 WHERE id = $2",
		analysisJSON,
		id,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save meeting analysis",
		})

		return
	}

	// ========================================================
	// STEP 5: COMPLETED
	// ========================================================

	_, err = h.DB.Exec(
		context.Background(),
		"UPDATE meetings SET status = $1 WHERE id = $2",
		"completed",
		id,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update meeting status",
		})

		return
	}

	// ========================================================
	// FINAL RESPONSE
	// ========================================================

	c.JSON(http.StatusOK, gin.H{
		"message":    "Meeting processed successfully",
		"meeting_id": id,
		"status":     "completed",
		"transcript": transcription.Transcript,
		"analysis":   analysis,
	})
}
