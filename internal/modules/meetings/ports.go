package meetings

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RepositoryPort interface {
	CreateMeeting(ctx context.Context, meeting *Meeting) error
	GetMeetingByID(ctx context.Context, id uuid.UUID) (*Meeting, error)
	ListMeetingsForUser(ctx context.Context, userID uuid.UUID) ([]Meeting, error)
	UpdateMeeting(ctx context.Context, meeting *Meeting) error
	DeleteMeeting(ctx context.Context, id uuid.UUID) error
	AddParticipants(ctx context.Context, participants []MeetingParticipant) error
	GetParticipant(ctx context.Context, meetingID, userID uuid.UUID) (*MeetingParticipant, error)
	ListParticipants(ctx context.Context, meetingID uuid.UUID) ([]MeetingParticipant, error)
	UpdateParticipant(ctx context.Context, participant *MeetingParticipant) error
	DeleteParticipant(ctx context.Context, meetingID, userID uuid.UUID) error
	AddMeetingResource(ctx context.Context, item *MeetingResource) error
	GetMeetingResource(ctx context.Context, meetingID, resourceID uuid.UUID) (*MeetingResource, error)
	ListMeetingResources(ctx context.Context, meetingID uuid.UUID) ([]MeetingResource, error)
	DeleteMeetingResource(ctx context.Context, meetingID, resourceID uuid.UUID) error
}

type UserDirectoryPort interface {
	Exists(ctx context.Context, userID uuid.UUID) (bool, error)
}

type ResourceDirectoryPort interface {
	Exists(ctx context.Context, resourceID uuid.UUID) (bool, error)
	IsAvailable(ctx context.Context, resourceID uuid.UUID, startAt, endAt time.Time) (bool, error)
}

type NotificationPort interface {
	NotifyParticipantsInvited(ctx context.Context, meeting *Meeting, participantIDs []uuid.UUID)
	NotifyMeetingUpdated(ctx context.Context, meeting *Meeting)
	NotifyAlternativeRequested(ctx context.Context, meeting *Meeting, participant *MeetingParticipant)
	NotifyResourceConflict(ctx context.Context, meeting *Meeting, resourceID uuid.UUID, startAt, endAt time.Time)
}

type ServicePort interface {
	CreateMeeting(ctx context.Context, command CreateMeetingCommand) (*Meeting, error)
	ListMeetings(ctx context.Context, query ListMeetingsQuery) ([]Meeting, error)
	GetMeeting(ctx context.Context, query GetMeetingQuery) (*Meeting, error)
	UpdateMeeting(ctx context.Context, command UpdateMeetingCommand) (*Meeting, error)
	DeleteMeeting(ctx context.Context, meetingID, organizerUserID uuid.UUID) error
	AddParticipants(ctx context.Context, command AddParticipantsCommand) (*Meeting, error)
	RemoveParticipant(ctx context.Context, command RemoveParticipantCommand) error
	AddResources(ctx context.Context, command AddResourceCommand) (*Meeting, error)
	RemoveResource(ctx context.Context, command RemoveResourceCommand) error
	RespondInvitation(ctx context.Context, command RespondInvitationCommand) (*MeetingParticipant, error)
	RequestAlternative(ctx context.Context, command RequestAlternativeCommand) (*MeetingParticipant, error)
}
