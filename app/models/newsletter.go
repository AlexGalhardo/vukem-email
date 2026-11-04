package models

import "github.com/AlexGalhardo/vukem/orm"

type Project struct {
	orm.Model
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Domain      string `json:"domain"`
	FromEmail   string `json:"from_email"`
	Description string `json:"description"`
}

type Subscriber struct {
	orm.Model
	ProjectID       int64  `json:"project_id"`
	Email           string `json:"email"`
	Status          string `json:"status"`
	TokenHash       string `json:"-"`
	ExpiresAt       int64  `json:"-"`
	LastRequestedAt int64  `json:"-"`
}

type Campaign struct {
	orm.Model
	ProjectID  int64  `json:"project_id"`
	Subject    string `json:"subject"`
	Mode       string `json:"mode"`
	Source     string `json:"source"`
	BodyHTML   string `json:"-"`
	Status     string `json:"status"`
	Recipients int64  `json:"recipients"`
}

type Delivery struct {
	orm.Model
	CampaignID   int64  `json:"campaign_id"`
	SubscriberID int64  `json:"subscriber_id"`
	Status       string `json:"status"`
}
