package unit

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"go-test/db-utils/models"
	"go-test/middleware"
	"go-test/routers"
	"go-test/test/mocks"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestGetAnimals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// mock database implementation
	mockRepository := new(mocks.MockRepository)
	mockRepository.On("FindAll").Return([]models.Animal{
		{ID: 1, Name: "Lion", Type: 3, Description: "King of the jungle"},
		{ID: 2, Name: "Eagle", Type: 3, Description: "Majestic bird"},
	}, nil)

	// create an engine instance
	r := gin.Default()
	// pass mock repository to routers
	var mu sync.Mutex
	r.Use(middleware.ApiMiddleware(mu, mockRepository))
	r.GET("/", routers.GetAnimals)

	// prepare a testing request
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// perform a request
	r.ServeHTTP(w, req)

	// check correct serving
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `[{"id":1,"data":{"name":"Lion","type":3,"description":"King of the jungle"}},{"id":2,"data":{"name":"Eagle","type":3,"description":"Majestic bird"}}]`, w.Body.String())

	// ensure that indeed called all mock methods
	mockRepository.AssertExpectations(t)
}
