package eventimport

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type Preview struct {
	Provider        string `json:"provider"`
	SourceURL       string `json:"source_url"`
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	CoverURL        string `json:"cover_url,omitempty"`
	StartsAt        string `json:"starts_at"`
	EndsAt          string `json:"ends_at,omitempty"`
	LocationName    string `json:"location_name,omitempty"`
	LocationAddress string `json:"location_address,omitempty"`
}

type Provider interface {
	Name() string
	Match(host string) bool
	Preview(ctx context.Context, sourceURL string) (*Preview, error)
}

type Manager struct {
	providers []Provider
}

func NewManager(providers ...Provider) *Manager {
	return &Manager{providers: providers}
}

func DefaultManager() *Manager {
	return NewManager(NewLumaProvider())
}

func (m *Manager) Preview(ctx context.Context, rawURL string) (*Preview, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("URL must start with http:// or https://")
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid URL host")
	}
	host := strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))

	for _, p := range m.providers {
		if !p.Match(host) {
			continue
		}
		preview, err := p.Preview(ctx, rawURL)
		if err != nil {
			return nil, err
		}
		preview.Provider = p.Name()
		preview.SourceURL = rawURL
		return preview, nil
	}

	return nil, fmt.Errorf("unsupported event provider")
}
