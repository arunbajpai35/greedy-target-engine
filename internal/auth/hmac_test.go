package auth

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const secret = "test-secret"

func signedReq(t *testing.T, ts time.Time, secret string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/v1/delivery?app=foo&country=us&os=android", nil)
	tsStr := strconv.FormatInt(ts.Unix(), 10)
	r.Header.Set("X-Timestamp", tsStr)
	r.Header.Set("X-Signature", Sign(secret, tsStr, http.MethodGet, r.URL.Path, r.URL.Query()))
	return r
}

func handler200() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireHMAC_Disabled(t *testing.T) {
	mw := RequireHMAC("")
	r := httptest.NewRequest(http.MethodGet, "/v1/delivery?app=foo", nil)
	w := httptest.NewRecorder()
	mw(handler200()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireHMAC_ValidSignature(t *testing.T) {
	mw := RequireHMAC(secret)
	r := signedReq(t, time.Now(), secret)
	w := httptest.NewRecorder()
	mw(handler200()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireHMAC_MissingHeaders(t *testing.T) {
	mw := RequireHMAC(secret)
	r := httptest.NewRequest(http.MethodGet, "/v1/delivery?app=foo", nil)
	w := httptest.NewRecorder()
	mw(handler200()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing")
}

func TestRequireHMAC_WrongSecret(t *testing.T) {
	mw := RequireHMAC(secret)
	r := signedReq(t, time.Now(), "different-secret")
	w := httptest.NewRecorder()
	mw(handler200()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "signature mismatch")
}

func TestRequireHMAC_StaleTimestamp(t *testing.T) {
	mw := RequireHMAC(secret)
	r := signedReq(t, time.Now().Add(-10*time.Minute), secret)
	w := httptest.NewRecorder()
	mw(handler200()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "outside allowed window")
}

func TestRequireHMAC_QueryOrderInvariant(t *testing.T) {
	mw := RequireHMAC(secret)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	r := httptest.NewRequest(http.MethodGet, "/v1/delivery?os=android&country=us&app=foo", nil)
	r.Header.Set("X-Timestamp", ts)
	r.Header.Set("X-Signature", Sign(secret, ts, http.MethodGet, "/v1/delivery", map[string][]string{
		"app": {"foo"}, "country": {"us"}, "os": {"android"},
	}))
	w := httptest.NewRecorder()
	mw(handler200()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, "signature should be order-invariant")
}
