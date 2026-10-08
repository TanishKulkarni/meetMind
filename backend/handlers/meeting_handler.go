package handlers

import (
	"context"
	"net/http"

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
