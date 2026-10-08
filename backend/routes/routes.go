package routes

import (
	"net/http"

	"meet-ai/backend/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func SetupRoutes(router *gin.Engine, db *pgx.Conn) {

	meetingHandler := handlers.NewMeetingHandler(db)

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Meet AI backend is running",
		})
	})

	router.POST("/api/meetings", meetingHandler.CreateMeeting)

	router.GET("/api/meetings", meetingHandler.GetMeetings)

	router.POST("/api/meetings/:id/upload", meetingHandler.UploadMeetingAudio)
}
