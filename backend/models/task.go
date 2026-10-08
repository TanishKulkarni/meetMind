package models

import "time"

type Task struct {
	ID        int       `json:"id"`
	MeetingID int       `json:"meeting_id"`
	Task      string    `json:"task"`
	Assignee  *string   `json:"assignee"`
	Deadline  *string   `json:"deadline"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
