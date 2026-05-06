package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MaxClockSkew bounds how stale a request timestamp can be. Five minutes is
// generous enough to tolerate clock drift but tight enough that captured
// requests can't be replayed indefinitely.
const MaxClockSkew = 5 * time.Minute

func RequireHMAC(secret string) func(http.Handler) http.Handler {
	if secret == "" {
		return func(next http.Handler) http.Handler { return next }
	}
	key := []byte(secret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := verify(r, key, time.Now()); err != "" {
				http.Error(w, err, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func verify(r *http.Request, key []byte, now time.Time) string {
	tsHeader := r.Header.Get("X-Timestamp")
	sigHeader := r.Header.Get("X-Signature")
	if tsHeader == "" || sigHeader == "" {
		return "missing X-Timestamp or X-Signature"
	}
	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return "invalid X-Timestamp"
	}
	if drift := now.Unix() - ts; drift > int64(MaxClockSkew/time.Second) || drift < -int64(MaxClockSkew/time.Second) {
		return "timestamp outside allowed window"
	}
	sig, err := hex.DecodeString(sigHeader)
	if err != nil {
		return "invalid X-Signature encoding"
	}
	want := sign(key, tsHeader, r.Method, r.URL.Path, r.URL.Query())
	if !hmac.Equal(sig, want) {
		return "signature mismatch"
	}
	return ""
}

// Sign builds the canonical payload and returns the hex signature. Exposed so
// clients (and tests) can produce matching signatures without re-deriving the
// payload format.
func Sign(secret, ts, method, path string, query map[string][]string) string {
	return hex.EncodeToString(sign([]byte(secret), ts, method, path, query))
}

func sign(key []byte, ts, method, path string, query map[string][]string) []byte {
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(ts)
	b.WriteByte('\n')
	b.WriteString(method)
	b.WriteByte('\n')
	b.WriteString(path)
	b.WriteByte('\n')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		vs := query[k]
		sort.Strings(vs)
		for j, v := range vs {
			if j > 0 {
				b.WriteByte('&')
			}
			b.WriteString(k)
			b.WriteByte('=')
			b.WriteString(v)
		}
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(b.String()))
	return mac.Sum(nil)
}
