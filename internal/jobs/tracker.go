package jobs

import (
	"sync"
	"time"
)

// Status represents the current state of a deploy job.
type Status string

const (
	StatusPending  Status = "pending"
	StatusCloning  Status = "cloning"
	StatusBuilding Status = "building"
	StatusStarting Status = "starting"
	StatusDone     Status = "done"
	StatusFailed   Status = "failed"
)

// DeployJob tracks the state of a project deployment.
type DeployJob struct {
	ProjectName string `json:"project_name"`
	GitURL      string `json:"git_url"`
	NetworkID   string `json:"network_id"`
	Status      Status `json:"status"`
	Error       string `json:"error,omitempty"`
	StartedAt   int64  `json:"started_at"`
	FinishedAt  int64  `json:"finished_at,omitempty"`
}

// Tracker manages deploy job state in memory.
type Tracker struct {
	mu   sync.RWMutex
	jobs map[string]*DeployJob
}

// NewTracker creates a new job tracker.
func NewTracker() *Tracker {
	return &Tracker{
		jobs: make(map[string]*DeployJob),
	}
}

// Create registers a new deploy job.
func (t *Tracker) Create(projectName, gitURL, networkID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.jobs[projectName] = &DeployJob{
		ProjectName: projectName,
		GitURL:      gitURL,
		NetworkID:   networkID,
		Status:      StatusPending,
		StartedAt:   time.Now().Unix(),
	}
}

// UpdateStatus sets the status of a running job.
func (t *Tracker) UpdateStatus(projectName string, status Status) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if job, ok := t.jobs[projectName]; ok {
		job.Status = status
		if status == StatusDone || status == StatusFailed {
			job.FinishedAt = time.Now().Unix()
		}
	}
}

// SetError marks a job as failed with an error message.
func (t *Tracker) SetError(projectName string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if job, ok := t.jobs[projectName]; ok {
		job.Status = StatusFailed
		job.Error = err.Error()
		job.FinishedAt = time.Now().Unix()
	}
}

// Get returns a copy of a deploy job. Returns nil if not found.
func (t *Tracker) Get(projectName string) *DeployJob {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if job, ok := t.jobs[projectName]; ok {
		copy := *job
		return &copy
	}
	return nil
}

// List returns a copy of all deploy jobs.
func (t *Tracker) List() []DeployJob {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]DeployJob, 0, len(t.jobs))
	for _, job := range t.jobs {
		result = append(result, *job)
	}
	return result
}
