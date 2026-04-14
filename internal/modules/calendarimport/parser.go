package calendarimport

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

type ICSParser struct{}

func NewICSParser() *ICSParser {
	return &ICSParser{}
}

func (p *ICSParser) Parse(content []byte) (*ParseResult, error) {
	calendar, err := ics.ParseCalendar(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse calendar: %w", err)
	}

	result := &ParseResult{
		Events: make([]ImportedEvent, 0, len(calendar.Events())),
	}
	for _, rawEvent := range calendar.Events() {
		importedEvent, skip, err := parseImportedEvent(rawEvent)
		if err != nil || skip {
			result.SkippedCount++
			continue
		}

		result.Events = append(result.Events, *importedEvent)
	}

	return result, nil
}

func parseImportedEvent(rawEvent *ics.VEvent) (*ImportedEvent, bool, error) {
	if rawEvent == nil {
		return nil, true, errors.New("event is nil")
	}
	if rawEvent.HasProperty(ics.ComponentPropertyRrule) ||
		rawEvent.HasProperty(ics.ComponentPropertyRDate) ||
		rawEvent.HasProperty(ics.ComponentPropertyExdate) ||
		rawEvent.HasProperty(ics.ComponentPropertyRecurrenceId) {
		return nil, true, errors.New("recurring events are not supported yet")
	}

	uid := strings.TrimSpace(propertyValue(rawEvent, ics.ComponentPropertyUniqueId))
	if uid == "" {
		return nil, true, errors.New("uid is required")
	}

	startAt, err := rawEvent.GetStartAt()
	if err != nil {
		return nil, true, fmt.Errorf("start_at is required: %w", err)
	}

	endAt, err := resolveEventEnd(rawEvent, startAt)
	if err != nil {
		return nil, true, err
	}
	if !startAt.Before(endAt) {
		return nil, true, errors.New("start_at must be before end_at")
	}

	title := strings.TrimSpace(propertyValue(rawEvent, ics.ComponentPropertySummary))
	if title == "" {
		title = "Imported event"
	}

	description := strings.TrimSpace(propertyValue(rawEvent, ics.ComponentPropertyDescription))
	visibilityHint := "busy"
	if strings.EqualFold(strings.TrimSpace(propertyValue(rawEvent, ics.ComponentPropertyTransp)), "TRANSPARENT") {
		visibilityHint = "free"
	}

	return &ImportedEvent{
		ExternalUID:     uid,
		Title:           title,
		Description:     description,
		StartAt:         startAt.UTC(),
		EndAt:           endAt.UTC(),
		Priority:        mapICSPriority(propertyValue(rawEvent, ics.ComponentPropertyPriority)),
		IsReschedulable: false,
		VisibilityHint:  visibilityHint,
	}, false, nil
}

func resolveEventEnd(rawEvent *ics.VEvent, startAt time.Time) (time.Time, error) {
	endAt, err := rawEvent.GetEndAt()
	if err == nil {
		return endAt, nil
	}

	startProp := rawEvent.GetProperty(ics.ComponentPropertyDtStart)
	if startProp != nil {
		if valueType, ok := startProp.ICalParameters[string(ics.ParameterValue)]; ok && len(valueType) > 0 && strings.EqualFold(valueType[0], "DATE") {
			return startAt.Add(24 * time.Hour), nil
		}
		if len(strings.TrimSpace(startProp.Value)) == 8 {
			return startAt.Add(24 * time.Hour), nil
		}
	}

	return time.Time{}, errors.New("end_at is required")
}

func propertyValue(rawEvent *ics.VEvent, property ics.ComponentProperty) string {
	item := rawEvent.GetProperty(property)
	if item == nil {
		return ""
	}

	return item.Value
}

func mapICSPriority(value string) string {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return "medium"
	}

	switch {
	case parsed >= 1 && parsed <= 2:
		return "critical"
	case parsed >= 3 && parsed <= 4:
		return "high"
	case parsed == 5:
		return "medium"
	case parsed >= 6:
		return "low"
	default:
		return "medium"
	}
}
