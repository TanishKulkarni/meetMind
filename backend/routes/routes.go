package routes

import (
	"net/http"

	"meet-ai/backend/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRoutes(router *gin.Engine, db *pgxpool.Pool) {

	meetingHandler := handlers.NewMeetingHandler(db)

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Meet AI backend is running",
		})
	})

	router.POST("/api/meetings", meetingHandler.CreateMeeting)

	router.GET("/api/meetings", meetingHandler.GetMeetings)

	router.GET("/api/meetings/:id", meetingHandler.GetMeeting)

	router.POST(
		"/api/meetings/:id/upload",
		meetingHandler.UploadMeetingAudio,
	)

	router.POST(
		"/api/meetings/:id/transcribe",
		meetingHandler.TranscribeMeeting,
	)

	router.POST(
		"/api/meetings/:id/analyze",
		meetingHandler.AnalyzeMeeting,
	)

	router.POST(
		"/api/meetings/:id/process",
		meetingHandler.ProcessMeeting,
	)

	router.GET("/api/tasks", meetingHandler.GetTasks)
	router.GET("/api/meetings/:id/tasks", meetingHandler.GetMeetingTasks)
	router.PATCH("/api/tasks/:id/status", meetingHandler.UpdateTaskStatus)
}
