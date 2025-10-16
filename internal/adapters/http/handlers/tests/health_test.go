package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/stretchr/testify/assert"
)

func TestNewHealthHandler(t *testing.T) {
	handler := handlers.NewHealthHandler()
	assert.NotNil(t, handler)
}

func TestHealthHandler_Ping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns healthy status", func(t *testing.T) {
		handler := handlers.NewHealthHandler()
		
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		
		handler.Ping(ctx)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["ok"].(bool))
	})

	t.Run("always returns 200 OK", func(t *testing.T) {
		handler := handlers.NewHealthHandler()
		
		// Call multiple times to ensure consistency
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			
			handler.Ping(ctx)
			
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})
}
