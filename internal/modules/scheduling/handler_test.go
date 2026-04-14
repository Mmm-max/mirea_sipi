package scheduling

import (
	"context"
	"net/http"
	"testing"

	"sipi/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeSchedulingService struct {
	search func(ctx context.Context, command SearchSlotsCommand) (*SearchResult, error)
}

func (f fakeSchedulingService) SearchSlots(ctx context.Context, command SearchSlotsCommand) (*SearchResult, error) {
	return f.search(ctx, command)
}
func (f fakeSchedulingService) GetSlots(context.Context, GetSlotsQuery) (*SearchResult, error) {
	return &SearchResult{}, nil
}
func (f fakeSchedulingService) SelectSlot(context.Context, SelectSlotCommand) (*MeetingAggregate, error) {
	return nil, nil
}

func TestHandlerSearchSlots(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	meetingID := uuid.New()
	service := fakeSchedulingService{
		search: func(_ context.Context, command SearchSlotsCommand) (*SearchResult, error) {
			if command.MeetingID != meetingID || command.OrganizerUserID != userID || command.TopN != 3 {
				t.Fatalf("unexpected command: %+v", command)
			}
			return &SearchResult{MeetingID: command.MeetingID}, nil
		},
	}

	router := gin.New()
	router.Use(testutil.TestAuthMiddleware(userID))
	handler := NewHandler(service)
	router.POST("/meetings/:id/search-slots", handler.SearchSlots)

	rec := testutil.PerformJSONRequest(router, http.MethodPost, "/meetings/"+meetingID.String()+"/search-slots", []byte(`{"top_n":3}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
}
