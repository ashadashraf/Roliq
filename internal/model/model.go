package model

import "time"

type Onboarding struct {
	CurrentStep   int        `json:"currentStep"`
	Status        string     `json:"status"`
	ProfileMethod *string    `json:"profileMethod,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
}

type Session struct {
	User         User         `json:"user"`
	Organization Organization `json:"organization"`
	Onboarding   Onboarding   `json:"onboarding"`
}

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}

type CareerProfile struct {
	Headline        string       `json:"headline"`
	Summary         string       `json:"summary"`
	CountryCode     string       `json:"countryCode"`
	TimeZone        string       `json:"timeZone"`
	City            string       `json:"city"`
	YearsExperience *float64     `json:"yearsExperience,omitempty"`
	Skills          []string     `json:"skills"`
	Experiences     []Experience `json:"experiences"`
	Education       []Education  `json:"education"`
	UpdatedAt       *time.Time   `json:"updatedAt,omitempty"`
}

type Experience struct {
	ID          string  `json:"id,omitempty"`
	Company     string  `json:"company"`
	Title       string  `json:"title"`
	Location    string  `json:"location,omitempty"`
	StartDate   string  `json:"startDate"`
	EndDate     *string `json:"endDate,omitempty"`
	IsCurrent   bool    `json:"isCurrent"`
	Description string  `json:"description,omitempty"`
}

type Education struct {
	ID           string  `json:"id,omitempty"`
	Institution  string  `json:"institution"`
	Degree       string  `json:"degree,omitempty"`
	FieldOfStudy string  `json:"fieldOfStudy,omitempty"`
	StartDate    *string `json:"startDate,omitempty"`
	EndDate      *string `json:"endDate,omitempty"`
}

type Resume struct {
	ID              string    `json:"id"`
	FileName        string    `json:"fileName"`
	Status          string    `json:"status"`
	RejectionReason *string   `json:"rejectionReason,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Dashboard struct {
	ProfileCompletion int            `json:"profileCompletion"`
	Onboarding        Onboarding     `json:"onboarding"`
	Profile           *CareerProfile `json:"profile"`
	Resumes           []Resume       `json:"resumes"`
}
