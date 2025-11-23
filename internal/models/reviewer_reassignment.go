package models

import "github.com/google/uuid"

type ReviewerReassignment struct {
	PRID         string
	OldReviewerID uuid.UUID
	NewReviewerID uuid.UUID
}

