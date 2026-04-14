package calendarimport

import "time"

type ImportJobResponse struct {
	ID               string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID           string  `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OriginalFilename string  `json:"original_filename" example:"work-calendar.ics"`
	Status           string  `json:"status" example:"completed"`
	ImportedCount    int     `json:"imported_count" example:"12"`
	UpdatedCount     int     `json:"updated_count" example:"3"`
	SkippedCount     int     `json:"skipped_count" example:"1"`
	ErrorMessage     *string `json:"error_message,omitempty" example:"invalid iCalendar file"`
	CreatedAt        string  `json:"created_at" example:"2026-03-26T12:00:00Z"`
	UpdatedAt        string  `json:"updated_at" example:"2026-03-26T12:00:01Z"`
}

type ImportJobEnvelope struct {
	Success bool              `json:"success" example:"true"`
	Data    ImportJobResponse `json:"data"`
	Meta    ResponseMetaDTO   `json:"meta"`
}

type ImportJobListResponse struct {
	Items []ImportJobResponse `json:"items"`
}

type ImportJobListEnvelope struct {
	Success bool                  `json:"success" example:"true"`
	Data    ImportJobListResponse `json:"data"`
	Meta    ResponseMetaDTO       `json:"meta"`
}

type ResponseMetaDTO struct {
	RequestID string `json:"request_id" example:"e7cc5d4a-3d85-4938-bff3-d2a6cf8ca2bb"`
}

func toImportJobResponse(job *ImportJob) ImportJobResponse {
	return ImportJobResponse{
		ID:               job.ID.String(),
		UserID:           job.UserID.String(),
		OriginalFilename: job.OriginalFilename,
		Status:           string(job.Status),
		ImportedCount:    job.ImportedCount,
		UpdatedCount:     job.UpdatedCount,
		SkippedCount:     job.SkippedCount,
		ErrorMessage:     job.ErrorMessage,
		CreatedAt:        job.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        job.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toImportJobResponses(jobs []ImportJob) []ImportJobResponse {
	result := make([]ImportJobResponse, 0, len(jobs))
	for i := range jobs {
		job := jobs[i]
		result = append(result, toImportJobResponse(&job))
	}

	return result
}
