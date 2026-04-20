package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"docker-visual/internal/config"
	"docker-visual/internal/jobs"
	"docker-visual/internal/models"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/gin-gonic/gin"
)

// mockDockerClient implements docker.DockerClient for testing.
type mockDockerClient struct {
	containers      []types.Container
	containerJSON   types.ContainerJSON
	networks        []network.Inspect
	networkInspect  network.Inspect
	images          []image.Summary
	volumes         []*volume.Volume
	pingErr         error
	listErr         error
	startErr        error
	stopErr         error
	removeErr       error
	createNetErr    error
	createdNetID    string
	runTunnelErr    error
}

func (m *mockDockerClient) Ping(ctx context.Context) error                        { return m.pingErr }
func (m *mockDockerClient) ListContainers(ctx context.Context) ([]types.Container, error) {
	return m.containers, m.listErr
}
func (m *mockDockerClient) GetContainer(ctx context.Context, id string) (types.ContainerJSON, error) {
	return m.containerJSON, m.listErr
}
func (m *mockDockerClient) StartContainer(ctx context.Context, id string) error   { return m.startErr }
func (m *mockDockerClient) StopContainer(ctx context.Context, id string) error    { return m.stopErr }
func (m *mockDockerClient) RemoveContainer(ctx context.Context, id string, force bool) error {
	return m.removeErr
}
func (m *mockDockerClient) ListNetworks(ctx context.Context) ([]network.Inspect, error) {
	return m.networks, m.listErr
}
func (m *mockDockerClient) GetNetwork(ctx context.Context, id string) (network.Inspect, error) {
	return m.networkInspect, m.listErr
}
func (m *mockDockerClient) ListImages(ctx context.Context) ([]image.Summary, error) {
	return m.images, m.listErr
}
func (m *mockDockerClient) ListVolumes(ctx context.Context) ([]*volume.Volume, error) {
	return m.volumes, m.listErr
}
func (m *mockDockerClient) CreateProjectNetwork(ctx context.Context, name string) (string, error) {
	return m.createdNetID, m.createNetErr
}
func (m *mockDockerClient) RunCloudflaredContainer(ctx context.Context, projectName, networkID, token string) error {
	return m.runTunnelErr
}
func (m *mockDockerClient) BuildImage(ctx context.Context, buildContextPath, imageName string) error {
	return nil
}
func (m *mockDockerClient) CreateAndStartContainer(ctx context.Context, imageName, networkID, projectName string) error {
	return nil
}
func (m *mockDockerClient) Close() error { return nil }

func setupRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	{
		api.GET("/health", h.Health)
		api.GET("/containers", h.ListContainers)
		api.GET("/containers/:id", h.GetContainer)
		api.POST("/containers/:id/start", h.StartContainer)
		api.POST("/containers/:id/stop", h.StopContainer)
		api.DELETE("/containers/:id", h.RemoveContainer)
		api.GET("/networks", h.ListNetworks)
		api.GET("/images", h.ListImages)
		api.GET("/volumes", h.ListVolumes)
		api.GET("/graph", h.GetGraphData)
		api.POST("/projects", h.CreateProject)
		api.GET("/projects", h.ListProjects)
		api.GET("/deploy/status/:name", h.GetDeployStatus)
		api.GET("/deploy/jobs", h.ListDeployJobs)
	}
	return r
}

func newTestHandler(mock *mockDockerClient) *Handler {
	cfg := &config.Config{
		WorkspacePath: "./test-workspaces",
	}
	return New(mock, cfg, jobs.NewTracker())
}

func TestHealth_OK(t *testing.T) {
	mock := &mockDockerClient{}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %q", body["status"])
	}
}

func TestHealth_DockerDown(t *testing.T) {
	mock := &mockDockerClient{pingErr: context.DeadlineExceeded}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestListContainers_Empty(t *testing.T) {
	mock := &mockDockerClient{containers: []types.Container{}}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/containers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var containers []models.Container
	json.Unmarshal(w.Body.Bytes(), &containers)
	if len(containers) != 0 {
		t.Errorf("expected 0 containers, got %d", len(containers))
	}
}

func TestListContainers_WithData(t *testing.T) {
	mock := &mockDockerClient{
		containers: []types.Container{
			{
				ID:     "abc123",
				Names:  []string{"/test-container"},
				Image:  "nginx:latest",
				State:  "running",
				Status: "Up 2 hours",
			},
		},
	}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/containers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var containers []models.Container
	json.Unmarshal(w.Body.Bytes(), &containers)
	if len(containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(containers))
	}
	if containers[0].ID != "abc123" {
		t.Errorf("expected id abc123, got %q", containers[0].ID)
	}
	if containers[0].Image != "nginx:latest" {
		t.Errorf("expected image nginx:latest, got %q", containers[0].Image)
	}
}

func TestStartContainer_InvalidID(t *testing.T) {
	mock := &mockDockerClient{}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("POST", "/api/containers/-invalid/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestStartContainer_OK(t *testing.T) {
	mock := &mockDockerClient{}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("POST", "/api/containers/abc123/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestStopContainer_OK(t *testing.T) {
	mock := &mockDockerClient{}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("POST", "/api/containers/abc123/stop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRemoveContainer_OK(t *testing.T) {
	mock := &mockDockerClient{}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("DELETE", "/api/containers/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateProject_InvalidName(t *testing.T) {
	mock := &mockDockerClient{}
	h := newTestHandler(mock)
	r := setupRouter(h)

	body := `{"name": "../../etc/passwd"}`
	req := httptest.NewRequest("POST", "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateProject_InvalidGitURL(t *testing.T) {
	mock := &mockDockerClient{}
	h := newTestHandler(mock)
	r := setupRouter(h)

	body := `{"name": "myproject", "gitUrl": "file:///etc/passwd"}`
	req := httptest.NewRequest("POST", "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateProject_OK_NoGit(t *testing.T) {
	mock := &mockDockerClient{createdNetID: "net-123"}
	h := newTestHandler(mock)
	r := setupRouter(h)

	body := `{"name": "myproject"}`
	req := httptest.NewRequest("POST", "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListProjects(t *testing.T) {
	mock := &mockDockerClient{
		networks: []network.Inspect{
			{
				ID:   "net-1",
				Name: "mynet",
				Labels: map[string]string{
					"docker-dashboard.project": "true",
					"docker-dashboard.name":    "myproject",
				},
				Containers: map[string]network.EndpointResource{
					"c1": {Name: "app"},
				},
			},
			{
				ID:   "net-2",
				Name: "bridge",
			},
		},
	}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var projects []models.Project
	json.Unmarshal(w.Body.Bytes(), &projects)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "myproject" {
		t.Errorf("expected name myproject, got %q", projects[0].Name)
	}
	if projects[0].Containers != 1 {
		t.Errorf("expected 1 container, got %d", projects[0].Containers)
	}
}

func TestDeployStatus_NotFound(t *testing.T) {
	mock := &mockDockerClient{}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/deploy/status/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestListImages_OK(t *testing.T) {
	mock := &mockDockerClient{
		images: []image.Summary{
			{ID: "img-1", Size: 1024, RepoTags: []string{"nginx:latest"}},
		},
	}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/images", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var images []models.Image
	json.Unmarshal(w.Body.Bytes(), &images)
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
	if images[0].RepoTags[0] != "nginx:latest" {
		t.Errorf("expected tag nginx:latest, got %q", images[0].RepoTags[0])
	}
}

func TestListVolumes_OK(t *testing.T) {
	mock := &mockDockerClient{
		volumes: []*volume.Volume{
			{Name: "vol-1", Driver: "local", Mountpoint: "/var/lib/docker/volumes/vol-1"},
		},
	}
	h := newTestHandler(mock)
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/volumes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var volumes []models.Volume
	json.Unmarshal(w.Body.Bytes(), &volumes)
	if len(volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(volumes))
	}
	if volumes[0].Name != "vol-1" {
		t.Errorf("expected name vol-1, got %q", volumes[0].Name)
	}
}
