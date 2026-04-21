package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"hop.top/kit/go/transport/api"

	"github.com/stretchr/testify/assert"
)

func TestChain_Order(t *testing.T) {
	var order []int
	mw := func(n int) api.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, n)
				next.ServeHTTP(w, r)
			})
		}
	}

	h := api.Chain(mw(1), mw(2), mw(3))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		order = append(order, 0)
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, []int{1, 2, 3, 0}, order)
}

func TestChain_Empty(t *testing.T) {
	called := false
	h := api.Chain()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	assert.True(t, called)
}

func TestWithMiddleware(t *testing.T) {
	called := false
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	r := api.NewRouter(api.WithMiddleware(mw))
	r.Handle("GET", "/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}
