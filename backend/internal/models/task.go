package models

import "time"

type Task struct {
	ID             int64      `json:"id"`
	CampaignID     *int64     `json:"campaign_id"`
	CategoryID     int64      `json:"category_id"`
	CategoryName   string     `json:"category_name"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Reward         float64    `json:"reward"`
	Slots          int        `json:"slots"`
	SlotsRemaining int        `json:"slots_remaining"`
	ApprovalRate   float64    `json:"approval_rate"`
	Status         string     `json:"status"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
}
