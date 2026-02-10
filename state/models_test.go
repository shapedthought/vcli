package state

import (
	"testing"
	"time"
)

func TestNewState(t *testing.T) {
	s := NewState()

	if s.Version != CurrentStateVersion {
		t.Errorf("Expected version %d, got %d", CurrentStateVersion, s.Version)
	}
	if s.Resources == nil {
		t.Fatal("Expected Resources map to be initialized")
	}
	if len(s.Resources) != 0 {
		t.Errorf("Expected empty Resources map, got %d entries", len(s.Resources))
	}
}

func TestStateSetAndGetResource(t *testing.T) {
	s := NewState()

	r := &Resource{
		Type: "VBRJob",
		ID:   "abc-123",
		Name: "TestJob",
		Spec: map[string]interface{}{"key": "value"},
	}

	s.SetResource(r)

	got, exists := s.GetResource("TestJob")
	if !exists {
		t.Fatal("Expected resource to exist after SetResource")
	}
	if got.ID != "abc-123" {
		t.Errorf("Expected ID=abc-123, got %s", got.ID)
	}
	if got.Type != "VBRJob" {
		t.Errorf("Expected Type=VBRJob, got %s", got.Type)
	}
}

func TestStateGetResourceMissing(t *testing.T) {
	s := NewState()

	_, exists := s.GetResource("nonexistent")
	if exists {
		t.Error("Expected exists=false for missing resource")
	}
}

func TestStateDeleteResource(t *testing.T) {
	s := NewState()
	s.SetResource(&Resource{Name: "ToDelete", Type: "VBRJob", ID: "1"})

	s.DeleteResource("ToDelete")

	_, exists := s.GetResource("ToDelete")
	if exists {
		t.Error("Expected resource to be deleted")
	}
}

func TestStateDeleteResourceMissing(t *testing.T) {
	s := NewState()

	// Should not panic
	s.DeleteResource("nonexistent")

	if len(s.Resources) != 0 {
		t.Errorf("Expected empty resources, got %d", len(s.Resources))
	}
}

func TestStateListResources(t *testing.T) {
	s := NewState()
	s.SetResource(&Resource{Name: "Job1", Type: "VBRJob", ID: "1"})
	s.SetResource(&Resource{Name: "Job2", Type: "VBRJob", ID: "2"})
	s.SetResource(&Resource{Name: "Repo1", Type: "VBRRepository", ID: "3"})

	jobs := s.ListResources("VBRJob")
	if len(jobs) != 2 {
		t.Errorf("Expected 2 VBRJob resources, got %d", len(jobs))
	}

	repos := s.ListResources("VBRRepository")
	if len(repos) != 1 {
		t.Errorf("Expected 1 VBRRepository resource, got %d", len(repos))
	}
}

func TestStateListResourcesNoMatch(t *testing.T) {
	s := NewState()
	s.SetResource(&Resource{Name: "Job1", Type: "VBRJob", ID: "1"})

	result := s.ListResources("VBREncryptionPassword")
	if len(result) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(result))
	}
}

func TestStateListResourcesEmptyType(t *testing.T) {
	s := NewState()
	s.SetResource(&Resource{Name: "Job1", Type: "VBRJob", ID: "1"})
	s.SetResource(&Resource{Name: "Repo1", Type: "VBRRepository", ID: "2"})

	// Empty type string returns all resources
	all := s.ListResources("")
	if len(all) != 2 {
		t.Errorf("Expected 2 resources for empty type filter, got %d", len(all))
	}
}

func TestAddEvent(t *testing.T) {
	r := &Resource{Name: "TestJob", Type: "VBRJob", ID: "1"}

	evt := ResourceEvent{
		Action:    "applied",
		Timestamp: time.Now(),
		User:      "admin",
	}
	r.AddEvent(evt)

	if len(r.History) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(r.History))
	}
	if r.History[0].Action != "applied" {
		t.Errorf("Expected action=applied, got %s", r.History[0].Action)
	}
}

func TestAddEventPrepends(t *testing.T) {
	r := &Resource{Name: "TestJob", Type: "VBRJob", ID: "1"}

	r.AddEvent(ResourceEvent{Action: "first", User: "admin"})
	r.AddEvent(ResourceEvent{Action: "second", User: "admin"})

	if len(r.History) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(r.History))
	}
	// Most recent should be first
	if r.History[0].Action != "second" {
		t.Errorf("Expected most recent event first, got %s", r.History[0].Action)
	}
	if r.History[1].Action != "first" {
		t.Errorf("Expected oldest event last, got %s", r.History[1].Action)
	}
}

func TestAddEventPrunesToMax(t *testing.T) {
	r := &Resource{Name: "TestJob", Type: "VBRJob", ID: "1"}

	// Add more than DefaultMaxHistoryEvents
	for i := 0; i < DefaultMaxHistoryEvents+5; i++ {
		r.AddEvent(ResourceEvent{
			Action: "applied",
			User:   "admin",
		})
	}

	if len(r.History) != DefaultMaxHistoryEvents {
		t.Errorf("Expected history pruned to %d, got %d", DefaultMaxHistoryEvents, len(r.History))
	}
}

func TestNewEvent(t *testing.T) {
	before := time.Now()
	evt := NewEvent("snapshotted", "admin")
	after := time.Now()

	if evt.Action != "snapshotted" {
		t.Errorf("Expected action=snapshotted, got %s", evt.Action)
	}
	if evt.User != "admin" {
		t.Errorf("Expected user=admin, got %s", evt.User)
	}
	if evt.Timestamp.Before(before) || evt.Timestamp.After(after) {
		t.Error("Expected timestamp to be between before and after")
	}
	if evt.Fields != nil {
		t.Errorf("Expected nil Fields, got %v", evt.Fields)
	}
	if evt.Partial {
		t.Error("Expected Partial=false")
	}
}

func TestNewEventWithFields(t *testing.T) {
	fields := []string{"name", "description"}
	evt := NewEventWithFields("applied", "admin", fields, true)

	if evt.Action != "applied" {
		t.Errorf("Expected action=applied, got %s", evt.Action)
	}
	if evt.User != "admin" {
		t.Errorf("Expected user=admin, got %s", evt.User)
	}
	if len(evt.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(evt.Fields))
	}
	if evt.Fields[0] != "name" || evt.Fields[1] != "description" {
		t.Errorf("Expected fields [name, description], got %v", evt.Fields)
	}
	if !evt.Partial {
		t.Error("Expected Partial=true")
	}
}
