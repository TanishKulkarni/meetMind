package models

import "time"

type Meeting struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	HostName    string      `json:"host_name"`
	HostEmail   string      `json:"host_email"`
	ScheduledAt time.Time   `json:"scheduled_at"`
	Status      string      `json:"status"`
	AudioPath   string      `json:"audio_path"`
	Transcript  string      `json:"transcript"`
	Analysis    interface{} `json:"analysis"`
	CreatedAt   time.Time   `json:"created_at"`
}
