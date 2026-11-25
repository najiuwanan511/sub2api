package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
)

// A fresh connection makes every probe include DNS, TCP, TLS (for HTTPS),
// and the wait for response headers. It never uses the account or system proxy.
var accountChannelProbeClient = &http.Client{
	Timeout: monitorPingTimeout,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: &http.Transport{
		DialContext:            safeDialContext,
		ForceAttemptHTTP2:      true,
		DisableKeepAlives:      true,
		MaxResponseHeaderBytes: 64 << 10,
	},
}

type AccountChannelProbeResult struct {
	Status     string    `json:"status"`
	LatencyMs  *int      `json:"latency_ms,omitempty"`
	HTTPStatus int       `json:"http_status,omitempty"`
	CheckedAt  time.Time `json:"checked_at"`
}

func (s *AccountTestService) ProbeChannelConnection(ctx context.Context, account *Account) AccountChannelProbeResult {
	return s.probeChannelConnection(ctx, account, accountChannelProbeClient)
}

func (s *AccountTestService) probeChannelConnection(ctx context.Context, account *Account, client *http.Client) (result AccountChannelProbeResult) {
	defer func() { result.CheckedAt = time.Now().UTC() }()
	result.Status = "unsupported"
	if account == nil || (account.Type != AccountTypeAPIKey && account.Type != AccountTypeUpstream) {
		return
	}
	base := strings.TrimSpace(account.GetCredential("base_url"))
	if base == "" {
		return
	}
	result.Status = "blocked"
	parsed, err := url.Parse(base)
	if err != nil || parsed.User != nil || parsed.ForceQuery || parsed.RawQuery != "" || strings.Contains(base, "#") {
		return
	}
	// Apply the gateway's URL policy, then additionally restrict this tool to
	// public addresses. safeDialContext connects to the IP it actually checked.
	if _, err := s.validateUpstreamBaseURL(base); err != nil {
		return
	}
	host := strings.TrimSuffix(parsed.Hostname(), ".")
	if isBlockedHostname(host) || urlvalidator.IsBlockedHost(host) || strings.Contains(host, "%") {
		return
	}
	if ip := net.ParseIP(host); ip != nil && isPrivateIP(ip) {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, parsed.String(), nil)
	if err != nil {
		return
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		result.Status = "unreachable"
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			result.Status = "timeout"
		}
		return
	}
	ms := int(time.Since(start).Milliseconds())
	_ = resp.Body.Close()
	result.Status = "reachable" // Any HTTP response establishes connectivity, not model health.
	result.LatencyMs = &ms
	result.HTTPStatus = resp.StatusCode
	return
}
