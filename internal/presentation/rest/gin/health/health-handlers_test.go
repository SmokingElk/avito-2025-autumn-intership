package healthhandlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	healthhandlers "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/health"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	expectedCode := http.StatusOK
	expectedBody := `{"name":"pull-request service","status":"ok"}`

	handlers := healthhandlers.CreateHealthHandlers()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", handlers.Health)

	req := httptest.NewRequest("GET", "/", nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, expectedCode, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}
