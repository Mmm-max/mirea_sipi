package scheduling

import (
	"context"
	"errors"

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

func (r *Repository) ReplaceSlots(ctx context.Context, meetingID uuid.UUID, slots []MeetingSlot) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var slotIDs []uuid.UUID
		if err := tx.Model(&MeetingSlotModel{}).Where("meeting_id = ?", meetingID).Pluck("id", &slotIDs).Error; err != nil {
			return err
		}
		if len(slotIDs) > 0 {
			if err := tx.Where("meeting_slot_id IN ?", slotIDs).Delete(&SlotConflictModel{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("meeting_id = ?", meetingID).Delete(&MeetingSlotModel{}).Error; err != nil {
			return err
		}
		for i := range slots {
			slotModel := toSlotModel(&slots[i])
			if err := tx.Create(slotModel).Error; err != nil {
				return err
			}
			for j := range slots[i].Conflicts {
				conflict := slots[i].Conflicts[j]
				conflict.MeetingSlotID = slotModel.ID
				conflictModel := toConflictModel(&conflict)
				if err := tx.Create(conflictModel).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *Repository) ListSlotsByMeetingID(ctx context.Context, meetingID uuid.UUID) ([]MeetingSlot, error) {
	var slotModels []MeetingSlotModel
	if err := r.db.WithContext(ctx).
		Where("meeting_id = ?", meetingID).
		Order("\"rank\" ASC, start_at ASC").
		Find(&slotModels).Error; err != nil {
		return nil, err
	}
	if len(slotModels) == 0 {
		return []MeetingSlot{}, nil
	}
	slotIDs := make([]uuid.UUID, 0, len(slotModels))
	for i := range slotModels {
		slotIDs = append(slotIDs, slotModels[i].ID)
	}
	var conflictModels []SlotConflictModel
	if err := r.db.WithContext(ctx).
		Where("meeting_slot_id IN ?", slotIDs).
		Order("created_at ASC").
		Find(&conflictModels).Error; err != nil {
		return nil, err
	}
	conflictsBySlot := make(map[uuid.UUID][]SlotConflict, len(slotModels))
	for i := range conflictModels {
		conflict := mapConflictModel(&conflictModels[i])
		conflictsBySlot[conflict.MeetingSlotID] = append(conflictsBySlot[conflict.MeetingSlotID], conflict)
	}
	result := make([]MeetingSlot, 0, len(slotModels))
	for i := range slotModels {
		slot := mapSlotModel(&slotModels[i])
		slot.Conflicts = conflictsBySlot[slot.ID]
		result = append(result, slot)
	}
	return result, nil
}

func (r *Repository) GetSlotByID(ctx context.Context, id uuid.UUID) (*MeetingSlot, error) {
	var slotModel MeetingSlotModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&slotModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	slot := mapSlotModel(&slotModel)
	var conflictModels []SlotConflictModel
	if err := r.db.WithContext(ctx).Where("meeting_slot_id = ?", id).Order("created_at ASC").Find(&conflictModels).Error; err != nil {
		return nil, err
	}
	for i := range conflictModels {
		slot.Conflicts = append(slot.Conflicts, mapConflictModel(&conflictModels[i]))
	}
	return &slot, nil
}

func toSlotModel(slot *MeetingSlot) *MeetingSlotModel {
	if slot == nil {
		return nil
	}
	return &MeetingSlotModel{
		BaseModel: db.BaseModel{
			ID:        slot.ID,
			CreatedAt: slot.CreatedAt,
			UpdatedAt: slot.UpdatedAt,
		},
		MeetingID: slot.MeetingID,
		StartAt:   slot.StartAt,
		EndAt:     slot.EndAt,
		Score:     slot.Score,
		Rank:      slot.Rank,
	}
}

func mapSlotModel(model *MeetingSlotModel) MeetingSlot {
	return MeetingSlot{
		ID:        model.ID,
		MeetingID: model.MeetingID,
		StartAt:   model.StartAt,
		EndAt:     model.EndAt,
		Score:     model.Score,
		Rank:      model.Rank,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func toConflictModel(conflict *SlotConflict) *SlotConflictModel {
	if conflict == nil {
		return nil
	}
	return &SlotConflictModel{
		BaseModel: db.BaseModel{
			ID:        conflict.ID,
			CreatedAt: conflict.CreatedAt,
			UpdatedAt: conflict.UpdatedAt,
		},
		MeetingSlotID:   conflict.MeetingSlotID,
		UserID:          conflict.UserID,
		EventID:         conflict.EventID,
		ResourceID:      conflict.ResourceID,
		ConflictType:    conflict.ConflictType,
		VisibleTitle:    conflict.VisibleTitle,
		VisiblePriority: conflict.VisiblePriority,
	}
}

func mapConflictModel(model *SlotConflictModel) SlotConflict {
	return SlotConflict{
		ID:              model.ID,
		MeetingSlotID:   model.MeetingSlotID,
		UserID:          model.UserID,
		EventID:         model.EventID,
		ResourceID:      model.ResourceID,
		ConflictType:    model.ConflictType,
		VisibleTitle:    model.VisibleTitle,
		VisiblePriority: model.VisiblePriority,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
