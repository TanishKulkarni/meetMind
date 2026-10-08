package handlers

import (
	"context"
	"fmt"
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
		          scheduled_at, status, audio_path, created_at
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
