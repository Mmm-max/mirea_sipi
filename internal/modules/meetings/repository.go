package meetings

import (
	"context"
	"errors"
	"strings"
	"time"

	"sipi/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateMeeting(ctx context.Context, meeting *Meeting) error {
	model := meetingToModel(meeting)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	fresh, err := r.GetMeetingByID(ctx, model.ID)
	if err != nil {
		return err
	}
	*meeting = *fresh
	return nil
}

func (r *Repository) GetMeetingByID(ctx context.Context, id uuid.UUID) (*Meeting, error) {
	var model MeetingModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	meeting := mapMeetingModel(&model)
	participants, err := r.ListParticipants(ctx, id)
	if err != nil {
		return nil, err
	}
	resources, err := r.ListMeetingResources(ctx, id)
	if err != nil {
		return nil, err
	}
	meeting.Participants = participants
	meeting.Resources = resources
	return &meeting, nil
}

func (r *Repository) ListMeetingsForUser(ctx context.Context, userID uuid.UUID) ([]Meeting, error) {
	var models []MeetingModel
	if err := r.db.WithContext(ctx).
		Model(&MeetingModel{}).
		Joins("LEFT JOIN meeting_participants ON meeting_participants.meeting_id = meetings.id").
		Where("meetings.organizer_user_id = ? OR meeting_participants.user_id = ?", userID, userID).
		Distinct("meetings.*").
		Order("meetings.created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]Meeting, 0, len(models))
	for i := range models {
		meeting, err := r.GetMeetingByID(ctx, models[i].ID)
		if err != nil {
			return nil, err
		}
		result = append(result, *meeting)
	}
	return result, nil
}

func (r *Repository) UpdateMeeting(ctx context.Context, meeting *Meeting) error {
	model := meetingToModel(meeting)
	if err := r.db.WithContext(ctx).
		Model(&MeetingModel{}).
		Where("id = ?", meeting.ID).
		Updates(map[string]any{
			"title":               model.Title,
			"description":         model.Description,
			"duration_minutes":    model.DurationMinutes,
			"priority":            model.Priority,
			"status":              model.Status,
			"search_range_start":  model.SearchRangeStart,
			"search_range_end":    model.SearchRangeEnd,
			"earliest_start_time": model.EarliestStartTime,
			"latest_start_time":   model.LatestStartTime,
			"recurrence_rule":     model.RecurrenceRule,
			"selected_start_at":   model.SelectedStartAt,
			"selected_end_at":     model.SelectedEndAt,
		}).Error; err != nil {
		return err
	}
	fresh, err := r.GetMeetingByID(ctx, meeting.ID)
	if err != nil {
		return err
	}
	*meeting = *fresh
	return nil
}

func (r *Repository) DeleteMeeting(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&MeetingModel{}, "id = ?", id).Error
}

func (r *Repository) AddParticipants(ctx context.Context, participants []MeetingParticipant) error {
	if len(participants) == 0 {
		return nil
	}
	models := make([]MeetingParticipantModel, 0, len(participants))
	for i := range participants {
		models = append(models, *participantToModel(&participants[i]))
	}
	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *Repository) GetParticipant(ctx context.Context, meetingID, userID uuid.UUID) (*MeetingParticipant, error) {
	var model MeetingParticipantModel
	if err := r.db.WithContext(ctx).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	participant := mapParticipantModel(&model)
	return &participant, nil
}

func (r *Repository) ListParticipants(ctx context.Context, meetingID uuid.UUID) ([]MeetingParticipant, error) {
	var models []MeetingParticipantModel
	if err := r.db.WithContext(ctx).
		Where("meeting_id = ?", meetingID).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]MeetingParticipant, 0, len(models))
	for i := range models {
		result = append(result, mapParticipantModel(&models[i]))
	}
	return result, nil
}

func (r *Repository) UpdateParticipant(ctx context.Context, participant *MeetingParticipant) error {
	model := participantToModel(participant)
	if err := r.db.WithContext(ctx).
		Model(&MeetingParticipantModel{}).
		Where("id = ?", participant.ID).
		Updates(map[string]any{
			"response_status":      model.ResponseStatus,
			"visibility_override":  model.VisibilityOverride,
			"alternative_start_at": model.AlternativeStartAt,
			"alternative_end_at":   model.AlternativeEndAt,
			"alternative_comment":  model.AlternativeComment,
		}).Error; err != nil {
		return err
	}
	fresh, err := r.GetParticipant(ctx, participant.MeetingID, participant.UserID)
	if err != nil {
		return err
	}
	*participant = *fresh
	return nil
}

func (r *Repository) DeleteParticipant(ctx context.Context, meetingID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Delete(&MeetingParticipantModel{}).Error
}

func (r *Repository) AddMeetingResource(ctx context.Context, item *MeetingResource) error {
	model := meetingResourceToModel(item)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	*item = mapMeetingResourceModel(model)
	return nil
}

func (r *Repository) GetMeetingResource(ctx context.Context, meetingID, resourceID uuid.UUID) (*MeetingResource, error) {
	var model MeetingResourceModel
	if err := r.db.WithContext(ctx).
		Where("meeting_id = ? AND resource_id = ?", meetingID, resourceID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	item := mapMeetingResourceModel(&model)
	return &item, nil
}

func (r *Repository) ListMeetingResources(ctx context.Context, meetingID uuid.UUID) ([]MeetingResource, error) {
	var models []MeetingResourceModel
	if err := r.db.WithContext(ctx).
		Where("meeting_id = ?", meetingID).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]MeetingResource, 0, len(models))
	for i := range models {
		result = append(result, mapMeetingResourceModel(&models[i]))
	}
	return result, nil
}

func (r *Repository) DeleteMeetingResource(ctx context.Context, meetingID, resourceID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("meeting_id = ? AND resource_id = ?", meetingID, resourceID).
		Delete(&MeetingResourceModel{}).Error
}

func meetingToModel(meeting *Meeting) *MeetingModel {
	if meeting == nil {
		return nil
	}
	model := &MeetingModel{
		BaseModel: db.BaseModel{
			ID:        meeting.ID,
			CreatedAt: meeting.CreatedAt,
			UpdatedAt: meeting.UpdatedAt,
		},
		OrganizerUserID:   meeting.OrganizerUserID,
		Title:             meeting.Title,
		Description:       meeting.Description,
		DurationMinutes:   meeting.DurationMinutes,
		Priority:          meeting.Priority,
		Status:            string(meeting.Status),
		SearchRangeStart:  meeting.SearchRangeStart,
		SearchRangeEnd:    meeting.SearchRangeEnd,
		EarliestStartTime: meeting.EarliestStartTime,
		LatestStartTime:   meeting.LatestStartTime,
		RecurrenceRule:    meeting.RecurrenceRule,
		SelectedStartAt:   meeting.SelectedStartAt,
		SelectedEndAt:     meeting.SelectedEndAt,
	}
	if meeting.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{Time: *meeting.DeletedAt, Valid: true}
	}
	return model
}

func mapMeetingModel(model *MeetingModel) Meeting {
	var deletedAt *time.Time
	if model.DeletedAt.Valid {
		deletedAt = &model.DeletedAt.Time
	}
	return Meeting{
		ID:                model.ID,
		OrganizerUserID:   model.OrganizerUserID,
		Title:             model.Title,
		Description:       model.Description,
		DurationMinutes:   model.DurationMinutes,
		Priority:          model.Priority,
		Status:            MeetingStatus(model.Status),
		SearchRangeStart:  model.SearchRangeStart,
		SearchRangeEnd:    model.SearchRangeEnd,
		EarliestStartTime: model.EarliestStartTime,
		LatestStartTime:   model.LatestStartTime,
		RecurrenceRule:    model.RecurrenceRule,
		SelectedStartAt:   model.SelectedStartAt,
		SelectedEndAt:     model.SelectedEndAt,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
		DeletedAt:         deletedAt,
	}
}

func participantToModel(participant *MeetingParticipant) *MeetingParticipantModel {
	if participant == nil {
		return nil
	}
	return &MeetingParticipantModel{
		BaseModel: db.BaseModel{
			ID:        participant.ID,
			CreatedAt: participant.CreatedAt,
			UpdatedAt: participant.UpdatedAt,
		},
		MeetingID:          participant.MeetingID,
		UserID:             participant.UserID,
		ResponseStatus:     string(participant.ResponseStatus),
		VisibilityOverride: participant.VisibilityOverride,
		AlternativeStartAt: participant.AlternativeStartAt,
		AlternativeEndAt:   participant.AlternativeEndAt,
		AlternativeComment: participant.AlternativeComment,
	}
}

func mapParticipantModel(model *MeetingParticipantModel) MeetingParticipant {
	return MeetingParticipant{
		ID:                 model.ID,
		MeetingID:          model.MeetingID,
		UserID:             model.UserID,
		ResponseStatus:     InvitationStatus(model.ResponseStatus),
		VisibilityOverride: model.VisibilityOverride,
		AlternativeStartAt: model.AlternativeStartAt,
		AlternativeEndAt:   model.AlternativeEndAt,
		AlternativeComment: model.AlternativeComment,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

func meetingResourceToModel(item *MeetingResource) *MeetingResourceModel {
	if item == nil {
		return nil
	}
	return &MeetingResourceModel{
		BaseModel: db.BaseModel{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		},
		MeetingID:  item.MeetingID,
		ResourceID: item.ResourceID,
	}
}

func mapMeetingResourceModel(model *MeetingResourceModel) MeetingResource {
	return MeetingResource{
		ID:         model.ID,
		MeetingID:  model.MeetingID,
		ResourceID: model.ResourceID,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key value")
}
