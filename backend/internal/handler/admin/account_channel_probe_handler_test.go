package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAccountChannelProbeHandlerUsesOnlySavedURL(t *testing.T) {
	adminSvc := &availableModelsAdminService{
		stubAdminService: newStubAdminService(),
		account: service.Account{
			ID: 44, Type: service.AccountTypeAPIKey,
			Credentials: map[string]any{"base_url": "https://127.0.0.1", "api_key": "never-returned"},
		},
	}
	accountTestSvc := service.NewAccountTestService(nil, nil, nil, nil, nil, nil, &config.Config{}, nil)
	h := &AccountHandler{adminService: adminSvc, accountTestService: accountTestSvc}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/accounts/:id/channel-probe", h.ProbeChannel)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/accounts/44/channel-probe", strings.NewReader(`{"base_url":"https://override.example.com"}`)))
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"status":"blocked"`)
	require.NotContains(t, w.Body.String(), "127.0.0.1")
	require.NotContains(t, w.Body.String(), "never-returned")
}
