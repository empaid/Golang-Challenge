package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"main/queries"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockStore struct {
	getUsersFn   func(context.Context) ([]queries.GetUsersQueryRow, error)
	createUserFn func(context.Context, string, string, string, *string, []byte) (*queries.GetUsersQueryRow, error)
}

func (m *mockStore) GetUsers(ctx context.Context) ([]queries.GetUsersQueryRow, error) {
	return m.getUsersFn(ctx)
}

func (m *mockStore) CreateUser(ctx context.Context, username, email, userType string, nickname *string, password []byte) (*queries.GetUsersQueryRow, error) {
	return m.createUserFn(ctx, username, email, userType, nickname, password)
}

func (m *mockStore) PatchUser(context.Context, int, *string, *string, *string, *string, bool) (*queries.GetUsersQueryRow, error) {
	panic("TODO")
}
func (m *mockStore) CreateUserMessage(context.Context, int, string) (*queries.GetUserMessagesQueryRow, error) {
	panic("TODO")
}

func (m *mockStore) FetchUserFromEmail(context.Context, string) (*queries.GetUsersQueryRow, error) {
	panic("TODO")
}

func TestGetUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		stub       func(context.Context) ([]queries.GetUsersQueryRow, error)
		expectCode int
		validate   func(*testing.T, []byte)
	}{
		{
			name: "multiple users",
			stub: func(ctx context.Context) ([]queries.GetUsersQueryRow, error) {
				return []queries.GetUsersQueryRow{
					{ID: 1, Username: "hardik", Email: "hardik@gmail.com", UserType: "UTYPE_USER", MessageCount: 2},
					{ID: 2, Username: "purohit", Email: "purohit@gmail.com", UserType: "UTYPE_ADMIN"},
				}, nil
			},
			expectCode: http.StatusOK,
			validate: func(t *testing.T, body []byte) {
				var resp struct {
					Users []struct {
						ID           int    `json:"id"`
						Email        string `json:"email"`
						UserType     string `json:"userType"`
						Username     string `json:"username"`
						MessageCount int    `json:"messageCount"`
					}
				}
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Len(t, resp.Users, 2)
				assert.Equal(t, "hardik", resp.Users[0].Username)
				assert.Equal(t, 0, resp.Users[1].MessageCount)
			},
		},
		{
			name: "no users",
			stub: func(ctx context.Context) ([]queries.GetUsersQueryRow, error) {
				return []queries.GetUsersQueryRow{}, nil
			},
			expectCode: http.StatusOK,
			validate: func(t *testing.T, body []byte) {
				var resp struct {
					Users []interface{} `json:"users"`
				}
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Empty(t, resp.Users)
			},
		},
		{
			name: "database error",
			stub: func(ctx context.Context) ([]queries.GetUsersQueryRow, error) {
				return nil, errors.New("DB Failed")
			},
			expectCode: http.StatusBadRequest,
			validate: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "Failed to retrieve users")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			handler := &UserHandler{UserDB: &mockStore{getUsersFn: tc.stub}}
			handler.GetUsers(ctx)
			assert.Equal(t, tc.expectCode, rec.Code)
			tc.validate(t, rec.Body.Bytes())
		})
	}
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		stub       func(context.Context, string, string, string, *string, []byte) (*queries.GetUsersQueryRow, error)
		payload    string
		expectCode int
		validate   func(*testing.T, []byte)
	}{
		{
			name: "successful creation",
			stub: func(ctx context.Context, username, email, userType string, nickname *string, password []byte) (*queries.GetUsersQueryRow, error) {
				return &queries.GetUsersQueryRow{
					ID: 42, Username: username, Email: email,
					UserType: userType,
					Nickname: pgtype.Text{String: *nickname, Valid: true},
				}, nil
			},
			payload:    `{"username":"hardik","email":"hardikpurohit26@gmail.com","userType":"UTYPE_USER","nickname":"hardik_nic", "password":"test_pass"}`,
			expectCode: http.StatusOK,
			validate: func(t *testing.T, body []byte) {
				var resp struct {
					User struct {
						ID       int    `json:"id"`
						Email    string `json:"email"`
						UserType string `json:"userType"`
						Username string `json:"username"`
					} `json:"user"`
				}
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, 42, resp.User.ID)
			},
		},
		{
			name: "database error",
			stub: func(ctx context.Context, username, email, userType string, nickname *string, password []byte) (*queries.GetUsersQueryRow, error) {
				return nil, errors.New("unique key violation")
			},
			payload:    `{"username":"hardik","email":"hardik@gmail.com","userType":"UTYPE_ADMIN"}`,
			expectCode: http.StatusBadRequest,
			validate: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "Failed to create user")
			},
		},
		{
			name:       "invalid JSON",
			stub:       nil,
			payload:    `{"username":}`,
			expectCode: http.StatusBadRequest,
			validate: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "Bad Request")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(tc.payload))
			req.Header.Set("Content-Type", "application/json")
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			handler := &UserHandler{UserDB: &mockStore{createUserFn: tc.stub}}
			handler.CreateUser(ctx)
			assert.Equal(t, tc.expectCode, rec.Code)
			tc.validate(t, rec.Body.Bytes())
		})
	}
}
