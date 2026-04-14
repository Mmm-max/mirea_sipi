package calendarimport

import "github.com/google/uuid"

type ImportICSCommand struct {
	UserID           uuid.UUID
	OriginalFilename string
	Content          []byte
}

type ListImportJobsQuery struct {
	UserID uuid.UUID
}

type GetImportJobQuery struct {
	UserID uuid.UUID
	ID     uuid.UUID
}
