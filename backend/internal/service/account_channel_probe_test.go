package service

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestAccountChannelProbeSendsOnlyHeadAndDoesNotFollowRedirects(t *testing.T) {
	for _, code := range []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusFound} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodHead || r.URL.RequestURI() != "/prefix/v1" || r.ContentLength != 0 {
					t.Errorf("unexpected request: %s %s length %d", r.Method, r.URL, r.ContentLength)
				}
				if r.Header.Get("Authorization") != "" || r.Header.Get("X-API-Key") != "" || r.Header.Get("Cookie") != "" {
					t.Error("probe forwarded credentials")
				}
				if code == http.StatusFound {
					w.Header().Set("Location", "/must-not-follow")
				}
				w.WriteHeader(code)
			}))
			defer server.Close()
			client := *accountChannelProbeClient
			client.Transport = &http.Transport{
				DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
				},
			}
			defer client.CloseIdleConnections()
			svc := &AccountTestService{cfg: &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}}
			account := &Account{Type: AccountTypeAPIKey, Credentials: map[string]any{
				"base_url": "http://upstream.example.com/prefix/v1", "api_key": "not-sent",
			}}
			got := svc.probeChannelConnection(t.Context(), account, &client)
			require.Equal(t, "reachable", got.Status)
			require.Equal(t, code, got.HTTPStatus)
			require.NotNil(t, got.LatencyMs)
			require.False(t, got.CheckedAt.IsZero())
		})
	}
}

func TestAccountChannelProbeRejectsInvalidDestinations(t *testing.T) {
	svc := &AccountTestService{cfg: &config.Config{}}
	for _, base := range []string{
		"https://127.0.0.1", "https://169.254.169.254", "https://100.100.100.200", "https://[::1]",
		"https://224.0.0.1", "https://metadata.google.internal.", "https://nested.localhost.",
		"https://user:key@upstream.example.com", "https://upstream.example.com?api_key=test",
		"http://upstream.example.com", "file:///etc/passwd", "https://upstream.example.com:0",
	} {
		got := svc.probeChannelConnection(t.Context(), &Account{Type: AccountTypeAPIKey, Credentials: map[string]any{"base_url": base}}, nil)
		require.Equal(t, "blocked", got.Status, base)
		require.Nil(t, got.LatencyMs)
	}
	allowlisted := &AccountTestService{cfg: &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{
		Enabled: true, UpstreamHosts: []string{"allowed.example.com"},
	}}}}
	got := allowlisted.probeChannelConnection(t.Context(), &Account{Type: AccountTypeUpstream, Credentials: map[string]any{"base_url": "https://other.example.com"}}, nil)
	require.Equal(t, "blocked", got.Status)
	got = svc.probeChannelConnection(t.Context(), &Account{Type: AccountTypeOAuth}, nil)
	require.Equal(t, "unsupported", got.Status)
	got = svc.probeChannelConnection(t.Context(), &Account{Type: AccountTypeAPIKey}, nil)
	require.Equal(t, "unsupported", got.Status)
}

func TestAccountChannelProbeTimeoutIsNotReachability(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Millisecond, Transport: accountProbeTimeoutTransport{}}
	svc := &AccountTestService{cfg: &config.Config{}}
	got := svc.probeChannelConnection(t.Context(), &Account{Type: AccountTypeUpstream, Credentials: map[string]any{"base_url": "https://upstream.example.com"}}, client)
	require.Equal(t, "timeout", got.Status)
	require.Zero(t, got.HTTPStatus)
	require.Nil(t, got.LatencyMs)
}

type accountProbeTimeoutTransport struct{}

func (accountProbeTimeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}
