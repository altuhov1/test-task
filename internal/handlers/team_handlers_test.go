package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"test-task/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockTeamManager struct {
	mock.Mock
}

func (m *MockTeamManager) CreateTeam(ctx context.Context, team models.Team) (*models.Team, error) {
	args := m.Called(ctx, team)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamManager) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func TestHandler_AddTeam_Success(t *testing.T) {
	mockManager := new(MockTeamManager)
	handler := &Handler{TeamManag: mockManager}

	inputTeam := models.Team{
		TeamName: "backend",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}


	mockManager.On("CreateTeam", mock.Anything, inputTeam).
		Return(&inputTeam, nil)

	body, _ := json.Marshal(inputTeam)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()

	handler.AddTeam(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "team")
	teamData := response["team"].(map[string]interface{})
	assert.Equal(t, "backend", teamData["team_name"])

	mockManager.AssertCalled(t, "CreateTeam", mock.Anything, inputTeam)
}

func TestHandler_AddTeam_TeamExists(t *testing.T) {
	mockManager := new(MockTeamManager)
	handler := &Handler{TeamManag: mockManager}

	inputTeam := models.Team{
		TeamName: "backend",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	mockManager.On("CreateTeam", mock.Anything, inputTeam).
		Return(nil, models.ErrTeamExists)


	body, _ := json.Marshal(inputTeam)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()


	handler.AddTeam(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "TEAM_EXISTS", errorData["code"])
	assert.Equal(t, "team_name already exists", errorData["message"])
}

func TestHandler_AddTeam_InvalidJSON(t *testing.T) {
	mockManager := new(MockTeamManager)
	handler := &Handler{TeamManag: mockManager}

	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()

	handler.AddTeam(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_GetTeam_Success(t *testing.T) {
	mockManager := new(MockTeamManager)
	handler := &Handler{TeamManag: mockManager}

	expectedTeam := &models.Team{
		TeamName: "backend",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	mockManager.On("GetTeam", mock.Anything, "backend").
		Return(expectedTeam, nil)

	req := httptest.NewRequest("GET", "/team/get?team_name=backend", nil)
	rr := httptest.NewRecorder()

	handler.GetTeam(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.Team
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "backend", response.TeamName)
	assert.Len(t, response.Members, 1)
}

func TestHandler_GetTeam_NotFound(t *testing.T) {
	mockManager := new(MockTeamManager)
	handler := &Handler{TeamManag: mockManager}

	mockManager.On("GetTeam", mock.Anything, "nonexistent").
		Return(nil, models.ErrNotFound)

	req := httptest.NewRequest("GET", "/team/get?team_name=nonexistent", nil)
	rr := httptest.NewRecorder()

	handler.GetTeam(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errorData["code"])
	assert.Equal(t, "resource not found", errorData["message"])
}

func TestHandler_GetTeam_MissingTeamName(t *testing.T) {
	mockManager := new(MockTeamManager)
	handler := &Handler{TeamManag: mockManager}

	req := httptest.NewRequest("GET", "/team/get", nil)
	rr := httptest.NewRecorder()

	handler.GetTeam(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_AddTeam_WrongMethod(t *testing.T) {
	mockManager := new(MockTeamManager)
	handler := &Handler{TeamManag: mockManager}

	req := httptest.NewRequest("GET", "/team/add", nil)
	rr := httptest.NewRecorder()

	handler.AddTeam(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}