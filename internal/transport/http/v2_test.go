package httptransport

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/endpoints"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/service"
)

const testDBConnStr = "postgres://postgres:password@localhost:5432/targeting_db?sslmode=disable"

func newRouter(t *testing.T) (chi.Router, func()) {
	t.Helper()
	db, err := sql.Open("postgres", testDBConnStr)
	if err != nil {
		t.Skip("db not available")
	}
	if err := db.Ping(); err != nil {
		t.Skip("db unreachable")
	}
	svc := service.NewDeliveryService(db)
	eps := endpoints.Endpoints{Delivery: endpoints.MakeDeliveryEndpoint(svc)}
	r := chi.NewRouter()
	RegisterV2Routes(r, eps)
	return r, func() { db.Close() }
}

func TestV2Delivery_Match(t *testing.T) {
	r, cleanup := newRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/v2/delivery?app=com.gametion.ludokinggame&country=us&os=android", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "spotify")
	assert.Contains(t, w.Body.String(), "subwaysurfer")
	assert.NotContains(t, w.Body.String(), `"status"`, "status must not leak to the wire")
}

func TestV2Delivery_NoMatch(t *testing.T) {
	r, cleanup := newRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/v2/delivery?app=com.test&country=zz&os=web", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestV2Delivery_CaseInsensitive(t *testing.T) {
	r, cleanup := newRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/v2/delivery?app=COM.GAMETION.LUDOKINGGAME&country=US&os=ANDROID", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "spotify")
}
