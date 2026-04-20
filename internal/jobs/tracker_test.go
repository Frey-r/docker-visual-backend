package jobs

import (
	"errors"
	"testing"
)

func TestTracker_CreateAndGet(t *testing.T) {
	tr := NewTracker()
	tr.Create("myproject", "https://github.com/user/repo", "net-123")

	job := tr.Get("myproject")
	if job == nil {
		t.Fatal("expected job, got nil")
	}
	if job.Status != StatusPending {
		t.Errorf("expected pending, got %q", job.Status)
	}
	if job.ProjectName != "myproject" {
		t.Errorf("expected myproject, got %q", job.ProjectName)
	}
}

func TestTracker_GetNotFound(t *testing.T) {
	tr := NewTracker()
	job := tr.Get("nonexistent")
	if job != nil {
		t.Errorf("expected nil, got %+v", job)
	}
}

func TestTracker_UpdateStatus(t *testing.T) {
	tr := NewTracker()
	tr.Create("proj", "", "net")
	tr.UpdateStatus("proj", StatusBuilding)

	job := tr.Get("proj")
	if job.Status != StatusBuilding {
		t.Errorf("expected building, got %q", job.Status)
	}
}

func TestTracker_SetError(t *testing.T) {
	tr := NewTracker()
	tr.Create("proj", "", "net")
	tr.SetError("proj", errors.New("build failed"))

	job := tr.Get("proj")
	if job.Status != StatusFailed {
		t.Errorf("expected failed, got %q", job.Status)
	}
	if job.Error != "build failed" {
		t.Errorf("expected 'build failed', got %q", job.Error)
	}
	if job.FinishedAt == 0 {
		t.Error("expected FinishedAt to be set")
	}
}

func TestTracker_List(t *testing.T) {
	tr := NewTracker()
	tr.Create("proj1", "", "net1")
	tr.Create("proj2", "", "net2")

	list := tr.List()
	if len(list) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(list))
	}
}

func TestTracker_DoneStatus(t *testing.T) {
	tr := NewTracker()
	tr.Create("proj", "", "net")
	tr.UpdateStatus("proj", StatusDone)

	job := tr.Get("proj")
	if job.Status != StatusDone {
		t.Errorf("expected done, got %q", job.Status)
	}
	if job.FinishedAt == 0 {
		t.Error("expected FinishedAt to be set for done status")
	}
}
