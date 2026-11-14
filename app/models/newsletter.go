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
	ProjectID          int64  `json:"project_id"`
	Email              string `json:"email"`
	Status             string `json:"status"`
	TokenHash          string `json:"-"`
	ExpiresAt          int64  `json:"-"`
	LastRequestedAt    int64  `json:"-"`
	Confirmations      int64  `json:"-"`
	ConfirmationsSince int64  `json:"-"`
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

// Suppression blocks an address for one project after a hard bounce or spam complaint.
type Suppression struct {
	orm.Model
	ProjectID int64  `json:"project_id"`
	Email     string `json:"email"`
	Reason    string `json:"reason"`
}

// Bounce is one bounce or complaint reported by Resend. Type is "hard",
// "soft" or "complaint"; EventID (the Svix message id) makes a redelivered
// event idempotent.
type Bounce struct {
	orm.Model
	ProjectID    int64  `json:"project_id"`
	Email        string `json:"email"`
	SubscriberID int64  `json:"subscriber_id"`
	DeliveryID   int64  `json:"delivery_id"`
	Type         string `json:"type"`
	Reason       string `json:"reason"`
	EventID      string `json:"-"`
}

// DeliveryEvent is an open or click Resend reported for one delivery. Type is
// "opened" or "clicked"; URL is the clicked link without query or fragment.
type DeliveryEvent struct {
	orm.Model
	DeliveryID int64  `json:"delivery_id"`
	CampaignID int64  `json:"campaign_id"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	EventID    string `json:"-"`
}
