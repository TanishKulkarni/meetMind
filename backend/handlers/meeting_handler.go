package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"meet-ai/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MeetingHandler struct {
	DB *pgxpool.Pool
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

func NewMeetingHandler(db *pgxpool.Pool) *MeetingHandler {
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
		log.Printf("GetMeetings Query Error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
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
			log.Printf("GetMeetings Scan Error: %v", err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		meetings = append(meetings, meeting)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetMeetings Rows Error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid meeting ID",
		})
		return
	}

	// Atomically claim the meeting so concurrent requests cannot
	// process it at the same time. Failed meetings can be retried.
	var audioPath string
	err = h.DB.QueryRow(
		c.Request.Context(),
		`UPDATE meetings
         SET status = 'transcribing', error_message = NULL
         WHERE id = $1
           AND COALESCE(audio_path, '') <> ''
           AND status NOT IN ('transcribing', 'analyzing')
         RETURNING audio_path`,
		id,
	).Scan(&audioPath)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var status string
			var existingAudioPath string

			lookupErr := h.DB.QueryRow(
				c.Request.Context(),
				`SELECT status, COALESCE(audio_path, '')
                 FROM meetings WHERE id = $1`,
				id,
			).Scan(&status, &existingAudioPath)

			if errors.Is(lookupErr, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Meeting not found",
				})
				return
			}

			if lookupErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to check meeting status",
				})
				return
			}

			if existingAudioPath == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "No audio uploaded for this meeting",
				})
				return
			}

			if status == "transcribing" || status == "analyzing" {
				c.JSON(http.StatusConflict, gin.H{
					"error": "Meeting is already being processed",
				})
				return
			}

			c.JSON(http.StatusConflict, gin.H{
				"error": "Meeting could not be claimed for processing",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start meeting processing",
		})
		return
	}

	// Save the failure reason whenever a processing stage fails.
	fail := func(stage string, cause error) {
		_, dbErr := h.DB.Exec(
			context.Background(),
			`UPDATE meetings
             SET status = 'failed', error_message = $1
             WHERE id = $2`,
			stage+": "+cause.Error(),
			id,
		)

		if dbErr != nil {
			log.Printf(
				"Failed to persist processing error for meeting %d: %v",
				id, dbErr,
			)
		}

		log.Printf(
			"Meeting %d failed during %s: %v",
			id, stage, cause,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   stage + " failed",
			"details": cause.Error(),
		})
	}

	// 1. Transcribe the audio.
	transcription, err := h.transcribeAudio(audioPath)
	if err != nil {
		fail("Transcription", err)
		return
	}

	// 2. Save the transcript and advance the status.
	_, err = h.DB.Exec(
		c.Request.Context(),
		`UPDATE meetings
         SET transcript = $1, status = 'analyzing'
         WHERE id = $2`,
		transcription.Transcript,
		id,
	)
	if err != nil {
		fail("Saving transcript", err)
		return
	}

	// 3. Analyze the transcript with the existing AI service.
	analysis, err := h.analyzeTranscript(transcription.Transcript)
	if err != nil {
		fail("AI analysis", err)
		return
	}

	// 4. Serialize the analysis result.
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		fail("Encoding analysis", err)
		return
	}

	// 5. Save the analysis.
	_, err = h.DB.Exec(
		c.Request.Context(),
		`UPDATE meetings SET analysis = $1 WHERE id = $2`,
		analysisJSON,
		id,
	)
	if err != nil {
		fail("Saving analysis", err)
		return
	}

	// 6. Save extracted action items.
	if err := h.saveActionItems(id, analysis); err != nil {
		fail("Saving action items", err)
		return
	}

	// 7. Mark the meeting complete only after all stages succeed.
	_, err = h.DB.Exec(
		c.Request.Context(),
		`UPDATE meetings
         SET status = 'completed', error_message = NULL
         WHERE id = $1`,
		id,
	)
	if err != nil {
		fail("Completing meeting", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Meeting processed successfully",
		"meeting_id": id,
		"status":     "completed",
		"transcript": transcription.Transcript,
		"analysis":   analysis,
	})
}

func (h *MeetingHandler) saveActionItems(meetingID int, analysis interface{}) error {
	// Convert analysis to JSON.
	data, err := json.Marshal(analysis)
	if err != nil {
		return err
	}

	// Structure returned by Qwen.
	var result struct {
		ActionItems []struct {
			Task     string  `json:"task"`
			Assignee *string `json:"assignee"`
			Deadline *string `json:"deadline"`
		} `json:"action_items"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	// Start a database transaction.
	ctx := context.Background()

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Remove existing tasks inside the transaction.
	_, err = tx.Exec(
		ctx,
		`DELETE FROM tasks WHERE meeting_id = $1`,
		meetingID,
	)
	if err != nil {
		return err
	}

	// Insert all newly extracted action items.
	for _, item := range result.ActionItems {
		_, err = tx.Exec(
			ctx,
			`INSERT INTO tasks (
				meeting_id,
				task,
				assignee,
				deadline,
				status
			)
			VALUES ($1, $2, $3, $4, 'pending')`,
			meetingID,
			item.Task,
			item.Assignee,
			item.Deadline,
		)

		if err != nil {
			return err
		}
	}

	// Commit only after every insert succeeds.
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// GetTasks returns all tasks
func (h *MeetingHandler) GetTasks(c *gin.Context) {
	rows, err := h.DB.Query(
		context.Background(),
		`
		SELECT
			id,
			meeting_id,
			task,
			assignee,
			deadline,
			status,
			created_at
		FROM tasks
		ORDER BY created_at DESC
		`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch tasks",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		var task models.Task

		err := rows.Scan(
			&task.ID,
			&task.MeetingID,
			&task.Task,
			&task.Assignee,
			&task.Deadline,
			&task.Status,
			&task.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to read task",
				"details": err.Error(),
			})
			return
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed while reading tasks",
			"details": err.Error(),
		})
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}

// GetMeetingTasks returns tasks for a specific meeting
func (h *MeetingHandler) GetMeetingTasks(c *gin.Context) {
	id := c.Param("id")

	meetingID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid meeting ID",
		})
		return
	}

	rows, err := h.DB.Query(
		context.Background(),
		`
		SELECT
			id,
			meeting_id,
			task,
			assignee,
			deadline,
			status,
			created_at
		FROM tasks
		WHERE meeting_id = $1
		ORDER BY created_at DESC
		`,
		meetingID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch meeting tasks",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		var task models.Task

		err := rows.Scan(
			&task.ID,
			&task.MeetingID,
			&task.Task,
			&task.Assignee,
			&task.Deadline,
			&task.Status,
			&task.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to read task",
				"details": err.Error(),
			})
			return
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed while reading tasks",
			"details": err.Error(),
		})
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}

// UpdateTaskStatus updates the status of a task
func (h *MeetingHandler) UpdateTaskStatus(c *gin.Context) {
	id := c.Param("id")

	taskID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	var request struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if request.Status != "pending" && request.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status must be pending or completed",
		})
		return
	}

	var task models.Task

	err = h.DB.QueryRow(
		context.Background(),
		`
		UPDATE tasks
		SET status = $1
		WHERE id = $2
		RETURNING
			id,
			meeting_id,
			task,
			assignee,
			deadline,
			status,
			created_at
		`,
		request.Status,
		taskID,
	).Scan(
		&task.ID,
		&task.MeetingID,
		&task.Task,
		&task.Assignee,
		&task.Deadline,
		&task.Status,
		&task.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": task,
	})
}
