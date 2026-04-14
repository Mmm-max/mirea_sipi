package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"sipi/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeNotificationsService struct {
	list func(ctx context.Context, query ListNotificationsQuery) ([]Notification, error)
}

func (f fakeNotificationsService) CreateMany(context.Context, []CreateNotificationInput) error {
	return nil
}
func (f fakeNotificationsService) ListNotifications(ctx context.Context, query ListNotificationsQuery) ([]Notification, error) {
	return f.list(ctx, query)
}
func (f fakeNotificationsService) MarkRead(context.Context, MarkReadCommand) error { return nil }
func (f fakeNotificationsService) MarkAllRead(context.Context, MarkAllReadCommand) (int64, error) {
	return 0, nil
}

func TestHandlerListNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	service := fakeNotificationsService{
		list: func(_ context.Context, query ListNotificationsQuery) ([]Notification, error) {
			return []Notification{{
				ID:        uuid.New(),
				UserID:    query.UserID,
				Type:      NotificationTypeMeetingInvitation,
				Title:     "Invitation",
				Body:      "You have a meeting invitation",
				CreatedAt: time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC),
			}}, nil
		},
	}

	router := gin.New()
	router.Use(testutil.TestAuthMiddleware(userID))
	handler := NewHandler(service)
	router.GET("/notifications", handler.ListNotifications)

	rec := testutil.PerformJSONRequest(router, http.MethodGet, "/notifications", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Items []NotificationResponse `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !response.Success || len(response.Data.Items) != 1 {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}
