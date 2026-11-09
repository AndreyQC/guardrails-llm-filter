package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSelfCheck(t *testing.T) {
	t.Run("healthy endpoint returns nil", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)

		// httptest listens on 127.0.0.1:<port> — already loopback.
		require.NoError(t, RunSelfCheck(srv.URL[len("http://"):]))
	})

	t.Run("unhealthy endpoint returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		t.Cleanup(srv.Close)

		err := RunSelfCheck(srv.URL[len("http://"):])
		require.Error(t, err)
		assert.Contains(t, err.Error(), "503")
	})

	t.Run("unreachable endpoint returns error", func(t *testing.T) {
		// Port 1 on loopback is not served in practice.
		err := RunSelfCheck("127.0.0.1:1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "health probe")
	})
}

func TestLoopbackAddr(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{":8080", "127.0.0.1:8080"},
		{"0.0.0.0:9080", "127.0.0.1:9080"},
		{"[::]:8080", "127.0.0.1:8080"},
		{"127.0.0.1:8080", "127.0.0.1:8080"},
		{"[::1]:8080", "127.0.0.1:8080"},
		{"localhost:8080", "localhost:8080"},
		{"10.1.2.3:8080", "10.1.2.3:8080"},
		{"8080", "8080"},
	} {
		assert.Equal(t, tc.want, loopbackAddr(tc.in), "input %q", tc.in)
	}
}
