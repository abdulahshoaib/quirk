package handlers

import (
<<<<<<< Updated upstream
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	// "sync"
	"testing"
	"time"
)

func TestHandleResult_Completed(t *testing.T) {
	// Mock state
	id := "job1"
	jobStatuses[id] = JobStatus{Status: "completed", ETA: time.Now()}
	jobResults[id] = Result{Triples: []string{"a"}, Embeddings: [][]float64{{0.1, 0.2}}}

	req := httptest.NewRequest("GET", "/result?object_id="+id, nil)
	w := httptest.NewRecorder()

	HandleResult(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var result Result
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Fatal("Invalid JSON response")
	}
	if len(result.Embeddings) != 1 {
		t.Error("Expected one embedding")
	}
}

func TestHandleExport_JSON(t *testing.T) {
	id := "job2"
	jobResults[id] = Result{Triples: []string{"a"}, Embeddings: [][]float64{{0.1, 0.2}}}

	req := httptest.NewRequest("GET", "/export?object_id="+id+"&format=json", nil)
	w := httptest.NewRecorder()

	HandleExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Error("Expected application/json content type")
	}
}

func TestHandleExport_CSV(t *testing.T) {
	id := "job3"
	jobResults[id] = Result{Triples: []string{"hello"}, Embeddings: [][]float64{{1.1, 2.2}}}

	req := httptest.NewRequest("GET", "/export?object_id="+id+"&format=csv", nil)
	w := httptest.NewRecorder()

	HandleExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "hello") {
		t.Error("Expected triple in CSV export")
=======
	"bytes"
	// "context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	// "strings"
	// "sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDB is a mock implementation of sql.DB
type MockDB struct {
	mock.Mock
	sql.DB
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0).(sql.Result), arguments.Error(1)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	arguments := m.Called(query, args)
	return arguments.Get(0).(*sql.Row)
}

// MockResult is a mock implementation of sql.Result
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// MockRow is a mock implementation of sql.Row
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

func TestHandleSignup(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		dbExpectations func(*MockDB, *MockResult)
		expectedStatus int
	}{
		{
			name: "Successful signup",
			requestBody: map[string]string{
				"email": "test@example.com",
			},
			dbExpectations: func(db *MockDB, result *MockResult) {
				result.On("LastInsertId").Return(int64(1), nil)
				result.On("RowsAffected").Return(int64(1), nil)
				db.On("Exec", "INSERT INTO user_tokens (email, token) VALUES ($1, $2)", []interface{}{"test@example.com", mock.Anything}).Return(result, nil)
			},
			expectedStatus: http.StatusOK,
		},
		// ... other test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock DB
			mockDB := new(MockDB)
			mockResult := new(MockResult)
			tt.dbExpectations(mockDB, mockResult)

			// Also setup sqlmock for the actual DB calls
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			InitDB(db)

			// Set expectations for sqlmock
			mock.ExpectExec("INSERT INTO user_tokens").
				WithArgs("test@example.com", sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Call handler
			HandleSignup(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "token")
				assert.Contains(t, response, "email")
			}

			// Verify all expectations were met
			mockDB.AssertExpectations(t)
			mockResult.AssertExpectations(t)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthenticateJWT(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	InitDB(db)

	// Generate a valid token for testing
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": "test@example.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, _ := validToken.SignedString(jwtKey)

	tests := []struct {
		name           string
		authHeader     string
		mockExpect     func()
		expectedStatus int
	}{
		{
			name:       "Valid token",
			authHeader: "Bearer " + tokenString,
			mockExpect: func() {
				mock.ExpectQuery("SELECT token FROM user_tokens").
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"token"}).AddRow(tokenString))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing token",
			authHeader:     "",
			mockExpect:     func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Database error",
			authHeader: "Bearer " + tokenString,
			mockExpect: func() {
				mock.ExpectQuery("SELECT token FROM user_tokens").
					WithArgs("test@example.com").
					WillReturnError(fmt.Errorf("db error"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			AuthenticateJWT(nextHandler)(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHandleQuery(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Valid request",
			requestBody: map[string]interface{}{
				"req":  map[string]interface{}{},
				"text": []string{"sample text"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid request body",
			requestBody:    map[string]interface{}{"invalid": "data"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/query", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Call handler
			HandleQuery(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandleExport(t *testing.T) {
	// Setup test data
	testID := uuid.NewString()
	testResult := Result{
		Embeddings: [][]float64{{1.1, 2.2}, {3.3, 4.4}},
		Triples:    []string{"triple1", "triple2"},
	}

	// Add to jobResults
	mutex.Lock()
	jobResults[testID] = testResult
	mutex.Unlock()

	tests := []struct {
		name           string
		objectID       string
		format         string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "Export CSV",
			objectID:       testID,
			format:         "csv",
			expectedStatus: http.StatusOK,
			expectedHeader: "text/csv",
		},
		{
			name:           "Export JSON",
			objectID:       testID,
			format:         "json",
			expectedStatus: http.StatusOK,
			expectedHeader: "application/json",
		},
		{
			name:           "Missing object_id",
			objectID:       "",
			format:         "json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid format",
			objectID:       testID,
			format:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Result not found",
			objectID:       "nonexistent",
			format:         "json",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/export?object_id=%s&format=%s", tt.objectID, tt.format), nil)
			w := httptest.NewRecorder()

			// Call handler
			HandleExport(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Header().Get("Content-Type"), tt.expectedHeader)

				if tt.format == "csv" {
					reader := csv.NewReader(w.Body)
					records, err := reader.ReadAll()
					assert.NoError(t, err)
					assert.Greater(t, len(records), 1) // header + at least one row
				} else if tt.format == "json" {
					var result Result
					err := json.NewDecoder(w.Body).Decode(&result)
					assert.NoError(t, err)
					assert.Equal(t, testResult, result)
				}
			}
		})
	}
}

func TestHandleResult(t *testing.T) {
	// Setup test data
	completedID := uuid.NewString()
	inProgressID := uuid.NewString()

	mutex.Lock()
	jobStatuses[completedID] = JobStatus{Status: "completed"}
	jobResults[completedID] = Result{Embeddings: [][]float64{{1.0}}}

	jobStatuses[inProgressID] = JobStatus{Status: "in_progress"}
	mutex.Unlock()

	tests := []struct {
		name           string
		objectID       string
		expectedStatus int
	}{
		{
			name:           "Completed job",
			objectID:       completedID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "In progress job",
			objectID:       inProgressID,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Missing object_id",
			objectID:       "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Nonexistent job",
			objectID:       "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/result?object_id=%s", tt.objectID), nil)
			w := httptest.NewRecorder()

			// Call handler
			HandleResult(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var result Result
				err := json.NewDecoder(w.Body).Decode(&result)
				assert.NoError(t, err)
				assert.NotEmpty(t, result.Embeddings)
			}
		})
>>>>>>> Stashed changes
	}
}

func TestHandleStatus(t *testing.T) {
<<<<<<< Updated upstream
	id := "job4"
	jobStatuses[id] = JobStatus{Status: "processing", ETA: time.Now().Add(10 * time.Second)}

	req := httptest.NewRequest("GET", "/status?object_id="+id, nil)
	w := httptest.NewRecorder()

	HandleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "status") {
		t.Error("Expected status in response")
	}
}

func TestHandleResult_MissingID(t *testing.T) {
	req := httptest.NewRequest("GET", "/result", nil)
	w := httptest.NewRecorder()

	HandleResult(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing ID, got %d", w.Code)
	}
}

// Optional: Reset shared maps between tests
func TestMain(m *testing.M) {
	// clean state before tests
	mutex.Lock()
	jobStatuses = make(map[string]JobStatus)
	jobResults = make(map[string]Result)
	mutex.Unlock()

	m.Run()
=======
	// Setup test data
	testID := uuid.NewString()
	futureTime := time.Now().Add(10 * time.Second)

	mutex.Lock()
	jobStatuses[testID] = JobStatus{
		Status: "in_progress",
		ETA:    futureTime,
	}
	mutex.Unlock()

	tests := []struct {
		name           string
		objectID       string
		expectedStatus int
	}{
		{
			name:           "Valid status",
			objectID:       testID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing object_id",
			objectID:       "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Nonexistent job",
			objectID:       "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", fmt.Sprintf("/status?object_id=%s", tt.objectID), nil)
			w := httptest.NewRecorder()

			// Call handler
			HandleStatus(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response, "status")
				assert.Contains(t, response, "eta_seconds")
			}
		})
	}
}

func TestHandleProcess(t *testing.T) {
	// Helper to create multipart form data
	createMultipartForm := func(files map[string]string) (io.Reader, string) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		for filename, content := range files {
			part, _ := writer.CreateFormFile("files", filename)
			part.Write([]byte(content))
		}

		writer.Close()
		return body, writer.FormDataContentType()
	}

	tests := []struct {
		name           string
		files          map[string]string
		expectedStatus int
	}{
		{
			name: "Single file upload",
			files: map[string]string{
				"test1.txt": "file content 1",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Multiple files upload",
			files: map[string]string{
				"test1.txt": "file content 1",
				"test2.txt": "file content 2",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No files",
			files:          map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset job statuses and results
			mutex.Lock()
			jobStatuses = make(map[string]JobStatus)
			jobResults = make(map[string]Result)
			mutex.Unlock()

			// Create request with multipart form
			body, contentType := createMultipartForm(tt.files)
			req := httptest.NewRequest("POST", "/process", body)
			req.Header.Set("Content-Type", contentType)
			w := httptest.NewRecorder()

			// Call handler
			HandleProcess(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response, "object_id")

				// Verify job was created
				objectID := response["object_id"]
				mutex.RLock()
				status, exists := jobStatuses[objectID]
				mutex.RUnlock()
				assert.True(t, exists)
				assert.Equal(t, "in_progress", status.Status)
			}
		})
	}
}

func TestEmbeddingsToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []string
	}{
		{
			name:     "Single value",
			input:    []float64{1.23},
			expected: []string{"1.23"},
		},
		{
			name:     "Multiple values",
			input:    []float64{1.23, 4.56, 7.89},
			expected: []string{"1.23", "4.56", "7.89"},
		},
		{
			name:     "Empty slice",
			input:    []float64{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := embeddingsToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
>>>>>>> Stashed changes
}
