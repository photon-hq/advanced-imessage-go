// Package spectrumauth exchanges Spectrum project credentials for an Advanced
// iMessage gRPC server address and a self-refreshing [imessage.TokenSource]. It
// is shared by the example programs only; the library itself stays
// address-and-token-only, so this lives under examples/internal.
//
// The credential exchange calls the Spectrum REST API
// (POST /projects/{id}/imessage/tokens) with HTTP Basic auth and, for a
// dedicated project, derives the gRPC host from the selected instance id.
package spectrumauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	imessage "github.com/photon-hq/advanced-imessage-go"
)

// ProductionURL is the Spectrum REST API base for production projects.
const ProductionURL = "https://spectrum.photon.codes"

const (
	// imsgHostSuffix is appended to a dedicated instance id to form the gRPC
	// host:port. NOTE: this derivation is inferred from the SDK's documented
	// address format (e.g. "<instance>.imsg.photon.codes:443") and is not part
	// of the documented REST API — adjust it if Photon changes the scheme.
	imsgHostSuffix = ".imsg.photon.codes:443"

	// refreshSkew re-exchanges the token this long before it actually expires.
	refreshSkew = 60 * time.Second

	// httpTimeout bounds each Spectrum token-exchange request.
	httpTimeout = 30 * time.Second
)

// Config configures [Connect].
type Config struct {
	// ProjectID and ProjectSecret are the Spectrum project credentials.
	ProjectID     string
	ProjectSecret string
	// BaseURL overrides the Spectrum REST API base; empty uses [ProductionURL].
	BaseURL string
	// Line selects a dedicated line by phone number (optional).
	Line string
	// Instance selects a dedicated line by instance id (optional).
	Instance string
	// HTTPClient overrides the HTTP client used for the exchange (optional).
	HTTPClient *http.Client
}

// Line identifies the dedicated iMessage line that [Connect] selected.
type Line struct {
	// InstanceID is the dedicated instance id (the gRPC subdomain).
	InstanceID string
	// Number is the line's phone number.
	Number string
	// Address is the derived gRPC host:port to pass to [imessage.New].
	Address string
}

// Connect exchanges the project credentials, selects a dedicated line, derives
// its gRPC address, and returns a [imessage.TokenSource] that keeps the bearer
// token fresh across the client's lifetime.
func Connect(ctx context.Context, cfg Config) (imessage.TokenSource, Line, error) {
	if cfg.ProjectID == "" {
		return nil, Line{}, errors.New("spectrumauth: project ID is required")
	}
	if cfg.ProjectSecret == "" {
		return nil, Line{}, errors.New("spectrumauth: project secret is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = ProductionURL
	}
	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: httpTimeout}
	}

	tokens, err := exchange(ctx, hc, baseURL, cfg.ProjectID, cfg.ProjectSecret)
	if err != nil {
		return nil, Line{}, err
	}
	if tokens.Type != "dedicated" {
		return nil, Line{}, fmt.Errorf("spectrumauth: expected a dedicated project, but the exchange returned %q", tokens.Type)
	}

	instanceID, err := selectInstance(tokens, cfg.Instance, cfg.Line)
	if err != nil {
		return nil, Line{}, err
	}

	src := &tokenSource{
		projectID:     cfg.ProjectID,
		projectSecret: cfg.ProjectSecret,
		instanceID:    instanceID,
		baseURL:       baseURL,
		httpClient:    hc,
		token:         tokens.Auth[instanceID],
		expiry:        time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second).Add(-refreshSkew),
	}
	line := Line{
		InstanceID: instanceID,
		Number:     tokens.Numbers[instanceID],
		Address:    instanceID + imsgHostSuffix,
	}
	return src, line, nil
}

// tokenSource implements imessage.TokenSource by exchanging Spectrum project
// credentials for a dedicated instance's bearer token, caching it, and
// re-exchanging shortly before it expires. It is safe for concurrent use.
type tokenSource struct {
	projectID     string
	projectSecret string
	instanceID    string
	baseURL       string
	httpClient    *http.Client

	mu     sync.Mutex
	token  string
	expiry time.Time
}

var _ imessage.TokenSource = (*tokenSource)(nil)

// Token returns the cached bearer token for the selected instance, refreshing it
// via the Spectrum REST API when the cached one is near expiry.
func (s *tokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != "" && time.Now().Before(s.expiry) {
		return s.token, nil
	}

	tokens, err := exchange(ctx, s.httpClient, s.baseURL, s.projectID, s.projectSecret)
	if err != nil {
		return "", err
	}
	token, ok := tokens.Auth[s.instanceID]
	if !ok {
		return "", fmt.Errorf("spectrumauth: instance %q absent from token exchange response", s.instanceID)
	}
	s.token = token
	s.expiry = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second).Add(-refreshSkew)
	return s.token, nil
}

// tokens is the relevant subset of a dedicated token-exchange response.
type tokens struct {
	Type      string
	Auth      map[string]string // instance id -> bearer token
	Numbers   map[string]string // instance id -> phone number
	ExpiresIn int               // token lifetime in seconds
}

// exchange calls POST /projects/{projectId}/imessage/tokens with HTTP Basic auth
// (projectID:projectSecret) and returns the parsed credentials.
func exchange(ctx context.Context, hc *http.Client, baseURL, projectID, projectSecret string) (*tokens, error) {
	endpoint := strings.TrimRight(baseURL, "/") + "/projects/" + url.PathEscape(projectID) + "/imessage/tokens"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(projectID, projectSecret)
	req.Header.Set("Accept", "application/json")

	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("spectrumauth: token exchange: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var out struct {
		Succeed bool `json:"succeed"`
		Data    struct {
			Type      string            `json:"type"`
			Token     string            `json:"token"`   // shared projects
			Auth      map[string]string `json:"auth"`    // dedicated projects
			Numbers   map[string]string `json:"numbers"` // dedicated projects
			ExpiresIn int               `json:"expiresIn"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("spectrumauth: decoding token exchange response: %w", err)
	}
	if !out.Succeed {
		return nil, errors.New("spectrumauth: token exchange: server reported failure")
	}

	return &tokens{
		Type:      out.Data.Type,
		Auth:      out.Data.Auth,
		Numbers:   out.Data.Numbers,
		ExpiresIn: out.Data.ExpiresIn,
	}, nil
}

// selectInstance resolves which dedicated instance id to use, by explicit
// instance id, by phone number, or — when the project has exactly one line —
// automatically.
func selectInstance(t *tokens, wantInstance, wantLine string) (string, error) {
	if len(t.Auth) == 0 {
		return "", errors.New("spectrumauth: project has no dedicated iMessage lines")
	}
	switch {
	case wantInstance != "":
		if _, ok := t.Auth[wantInstance]; !ok {
			return "", fmt.Errorf("spectrumauth: instance %q not found; available lines: %s", wantInstance, describeLines(t))
		}
		return wantInstance, nil
	case wantLine != "":
		for id, number := range t.Numbers {
			if number == wantLine {
				return id, nil
			}
		}
		return "", fmt.Errorf("spectrumauth: no line with number %q; available lines: %s", wantLine, describeLines(t))
	case len(t.Auth) == 1:
		for id := range t.Auth {
			return id, nil
		}
	}
	return "", fmt.Errorf("spectrumauth: project has %d iMessage lines; pick one with Line or Instance: %s",
		len(t.Auth), describeLines(t))
}

// describeLines renders the available lines as "<instanceID> (<number>)" pairs,
// sorted for stable output.
func describeLines(t *tokens) string {
	ids := make([]string, 0, len(t.Auth))
	for id := range t.Auth {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if number := t.Numbers[id]; number != "" {
			parts = append(parts, fmt.Sprintf("%s (%s)", id, number))
		} else {
			parts = append(parts, id)
		}
	}
	return strings.Join(parts, ", ")
}
