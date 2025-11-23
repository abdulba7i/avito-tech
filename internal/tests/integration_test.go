package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reviewer-service/cmd/inits"
	"reviewer-service/internal/models"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	host := "localhost"
	port := "5434"
	user := "postgres"
	password := "postgres"
	dbname := "avito-tech"

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping test database: %v", err)
	}

	_, _ = db.Exec("TRUNCATE TABLE pr_reviewers, pull_requests, users, teams CASCADE")

	return db
}

func TestCreateTeamAndPR(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, inits.RunMigrations(db))

	services := inits.InitServices(db)
	handler := inits.SetupRoutes(services)

	teamReq := map[string]interface{}{
		"team_name": "test-team",
		"members": []map[string]interface{}{
			{"user_id": "11111111-1111-1111-1111-111111111111", "username": "User1", "is_active": true},
			{"user_id": "22222222-2222-2222-2222-222222222222", "username": "User2", "is_active": true},
			{"user_id": "33333333-3333-3333-3333-333333333333", "username": "User3", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(teamBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-test-1",
		"pull_request_name": "Test PR",
		"author_id":         "11111111-1111-1111-1111-111111111111",
	}
	prBody, _ := json.Marshal(prReq)
	req = httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(prBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var prResp struct {
		PR models.PullRequest `json:"pr"`
	}
	json.Unmarshal(w.Body.Bytes(), &prResp)

	assert.LessOrEqual(t, len(prResp.PR.Reviewers), 2)
	assert.NotContains(t, prResp.PR.Reviewers, "11111111-1111-1111-1111-111111111111")
	assert.Equal(t, "OPEN", string(prResp.PR.Status))
}

func TestMergePR(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, inits.RunMigrations(db))

	services := inits.InitServices(db)
	handler := inits.SetupRoutes(services)

	teamReq := map[string]interface{}{
		"team_name": "merge-team",
		"members": []map[string]interface{}{
			{"user_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "username": "Author", "is_active": true},
			{"user_id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "username": "Reviewer1", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(teamBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-merge-1",
		"pull_request_name": "Merge Test",
		"author_id":         "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
	}
	prBody, _ := json.Marshal(prReq)
	req = httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(prBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	mergeReq := map[string]interface{}{
		"pull_request_id": "pr-merge-1",
	}
	mergeBody, _ := json.Marshal(mergeReq)
	req = httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(mergeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var mergeResp struct {
		PR models.PullRequest `json:"pr"`
	}
	json.Unmarshal(w.Body.Bytes(), &mergeResp)
	assert.Equal(t, "MERGED", string(mergeResp.PR.Status))
	assert.NotNil(t, mergeResp.PR.MergedAt)

	req = httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(mergeBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReassignAfterMerge(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, inits.RunMigrations(db))

	services := inits.InitServices(db)
	handler := inits.SetupRoutes(services)

	teamReq := map[string]interface{}{
		"team_name": "reassign-team",
		"members": []map[string]interface{}{
			{"user_id": "cccccccc-cccc-cccc-cccc-cccccccccccc", "username": "Author2", "is_active": true},
			{"user_id": "dddddddd-dddd-dddd-dddd-dddddddddddd", "username": "Reviewer2", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(teamBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-reassign-1",
		"pull_request_name": "Reassign Test",
		"author_id":         "cccccccc-cccc-cccc-cccc-cccccccccccc",
	}
	prBody, _ := json.Marshal(prReq)
	req = httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(prBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	mergeReq := map[string]interface{}{"pull_request_id": "pr-reassign-1"}
	mergeBody, _ := json.Marshal(mergeReq)
	req = httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(mergeBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	reassignReq := map[string]interface{}{
		"pull_request_id": "pr-reassign-1",
		"old_user_id":     "dddddddd-dddd-dddd-dddd-dddddddddddd",
	}
	reassignBody, _ := json.Marshal(reassignReq)
	req = httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewReader(reassignBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var errResp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "PR_MERGED", errResp.Error.Code)
}

func TestStatistics(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	require.NoError(t, inits.RunMigrations(db))

	services := inits.InitServices(db)
	handler := inits.SetupRoutes(services)

	teamReq := map[string]interface{}{
		"team_name": "stats-team",
		"members": []map[string]interface{}{
			{"user_id": "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee", "username": "StatsUser1", "is_active": true},
			{"user_id": "ffffffff-ffff-ffff-ffff-ffffffffffff", "username": "StatsUser2", "is_active": true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(teamBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-stats-1",
		"pull_request_name": "Stats PR",
		"author_id":         "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
	}
	prBody, _ := json.Marshal(prReq)
	req = httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(prBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	req = httptest.NewRequest("GET", "/statistics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var statsResp struct {
		UserAssignments []struct {
			UserID string `json:"UserID"`
			Count  int    `json:"Count"`
		} `json:"user_assignments"`
		PRStats struct {
			TotalPRs int `json:"TotalPRs"`
		} `json:"pr_stats"`
	}
	json.Unmarshal(w.Body.Bytes(), &statsResp)

	assert.Greater(t, statsResp.PRStats.TotalPRs, 0)
}
