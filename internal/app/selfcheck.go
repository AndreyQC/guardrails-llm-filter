package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// selfCheckTimeout bounds the local probe: a healthy data plane answers in
// milliseconds, so anything slower is reported as a failure rather than
// blocking the container healthcheck.
const selfCheckTimeout = 3 * time.Second

// RunSelfCheck probes the local /healthz endpoint and returns an error when
// the data plane does not answer with 200 OK. It backs the `healthcheck`
// subcommand: the distroless runtime image has no shell or wget, so Docker's
// HEALTHCHECK execs the service binary, which probes itself over loopback.
func RunSelfCheck(listenAddr string) error {
	url := "http://" + loopbackAddr(listenAddr) + "/healthz"

	client := &http.Client{Timeout: selfCheckTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("health probe %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health probe %s: unexpected status %s", url, resp.Status)
	}
	return nil
}

// loopbackAddr rewrites a listen address (":8080", "0.0.0.0:8080",
// "[::]:8080", "10.0.0.1:8080") into one that always dials loopback, so the
// self-check never depends on external network interfaces.
func loopbackAddr(listenAddr string) string {
	// Bracketed IPv6 listener such as "[::]:8080".
	if rest, found := strings.CutPrefix(listenAddr, "["); found {
		if _, port, ok := strings.Cut(rest, "]:"); ok {
			return "127.0.0.1:" + port
		}
		return listenAddr
	}
	host, port, found := strings.Cut(listenAddr, ":")
	if !found {
		return listenAddr // no port separator: use as-is
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		return "127.0.0.1:" + port
	}
	return listenAddr
}
