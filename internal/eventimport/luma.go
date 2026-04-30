package eventimport

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type LumaProvider struct {
	client *http.Client
}

func NewLumaProvider() *LumaProvider {
	return &LumaProvider{
		client: &http.Client{Timeout: 12 * time.Second},
	}
}

func (p *LumaProvider) Name() string { return "luma" }

func (p *LumaProvider) Match(host string) bool {
	return host == "luma.com"
}

func (p *LumaProvider) Preview(ctx context.Context, sourceURL string) (*Preview, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "evermeet-importer/1.0")

	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch source failed")
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch source page")
	}
	bodyBytes, err := io.ReadAll(io.LimitReader(res.Body, 2_000_000))
	if err != nil {
		return nil, fmt.Errorf("read source failed")
	}
	body := string(bodyBytes)

	out := &Preview{
		Title:       metaContent(body, "og:title"),
		Description: metaContent(body, "og:description"),
		CoverURL:    firstNonEmpty(metaContent(body, "og:image"), metaContent(body, "og:image:url"), metaByName(body, "twitter:image"), metaByName(body, "image")),
	}
	if out.Title == "" {
		out.Title = titleTag(body)
	}
	out.Title = strings.TrimSpace(strings.TrimSuffix(out.Title, " · Luma"))
	if out.Description == "" {
		out.Description = metaByName(body, "description")
	}

	if ev := firstJSONLDEvent(body); ev != nil {
		if out.Title == "" {
			if v, _ := ev["name"].(string); v != "" {
				out.Title = v
			}
		}
		if out.Description == "" {
			if v, _ := ev["description"].(string); v != "" {
				out.Description = v
			}
		}
		if out.CoverURL == "" {
			out.CoverURL = imageValue(ev["image"])
		}
		if v, _ := ev["startDate"].(string); v != "" {
			out.StartsAt = normalizeDate(v)
		}
		if v, _ := ev["endDate"].(string); v != "" {
			out.EndsAt = normalizeDate(v)
		}
		if loc, _ := ev["location"].(map[string]any); loc != nil {
			if name, _ := loc["name"].(string); name != "" {
				out.LocationName = name
			}
			if addr, _ := loc["address"].(map[string]any); addr != nil {
				out.LocationAddress = addressString(addr)
			}
		}
	}

	if out.Title == "" || out.StartsAt == "" {
		return nil, fmt.Errorf("could not extract event title/start time from luma page")
	}
	out.CoverURL = resolveURL(sourceURL, out.CoverURL)
	out.CoverURL = unwrapLumaOGImage(out.CoverURL)
	return out, nil
}

func firstJSONLDEvent(body string) map[string]any {
	re := regexp.MustCompile(`(?is)<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	matches := re.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		raw := strings.TrimSpace(m[1])
		if raw == "" {
			continue
		}
		var parsed any
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			continue
		}
		if ev := findEventNode(parsed); ev != nil {
			return ev
		}
	}
	return nil
}

func findEventNode(v any) map[string]any {
	switch t := v.(type) {
	case map[string]any:
		if isEventType(t["@type"]) {
			return t
		}
		if g, ok := t["@graph"]; ok {
			if ev := findEventNode(g); ev != nil {
				return ev
			}
		}
		for _, value := range t {
			if ev := findEventNode(value); ev != nil {
				return ev
			}
		}
	case []any:
		for _, item := range t {
			if ev := findEventNode(item); ev != nil {
				return ev
			}
		}
	}
	return nil
}

func isEventType(v any) bool {
	switch t := v.(type) {
	case string:
		return strings.EqualFold(t, "Event")
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok && strings.EqualFold(s, "Event") {
				return true
			}
		}
	}
	return false
}

func metaContent(body, prop string) string {
	tags := metaTags(body)
	for _, tag := range tags {
		attr := parseTagAttrs(tag)
		if !strings.EqualFold(attr["property"], prop) {
			continue
		}
		if v := strings.TrimSpace(attr["content"]); v != "" {
			return v
		}
	}
	return ""
}

func metaByName(body, name string) string {
	tags := metaTags(body)
	for _, tag := range tags {
		attr := parseTagAttrs(tag)
		if !strings.EqualFold(attr["name"], name) {
			continue
		}
		if v := strings.TrimSpace(attr["content"]); v != "" {
			return v
		}
	}
	return ""
}

func titleTag(body string) string {
	re := regexp.MustCompile(`(?is)<title>(.*?)</title>`)
	if m := re.FindStringSubmatch(body); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func normalizeDate(v string) string {
	v = strings.TrimSpace(v)
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}
	return v
}

func addressString(addr map[string]any) string {
	parts := make([]string, 0, 4)
	for _, key := range []string{"streetAddress", "addressLocality", "addressRegion", "postalCode", "addressCountry"} {
		if v, _ := addr[key].(string); strings.TrimSpace(v) != "" {
			parts = append(parts, strings.TrimSpace(v))
		}
	}
	return strings.Join(parts, ", ")
}

func metaTags(body string) []string {
	re := regexp.MustCompile(`(?is)<meta\b[^>]*>`)
	return re.FindAllString(body, -1)
}

func parseTagAttrs(tag string) map[string]string {
	out := map[string]string{}
	re := regexp.MustCompile(`(?is)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*("([^"]*)"|'([^']*)')`)
	matches := re.FindAllStringSubmatch(tag, -1)
	for _, m := range matches {
		if len(m) < 5 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(m[1]))
		val := m[3]
		if val == "" {
			val = m[4]
		}
		out[key] = strings.TrimSpace(val)
	}
	return out
}

func imageValue(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []any:
		for _, item := range t {
			if out := imageValue(item); out != "" {
				return out
			}
		}
	case map[string]any:
		for _, key := range []string{"url", "contentUrl"} {
			if s, _ := t[key].(string); strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func resolveURL(baseURL, raw string) string {
	raw = strings.TrimSpace(html.UnescapeString(raw))
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.IsAbs() {
		return u.String()
	}
	b, err := url.Parse(baseURL)
	if err != nil {
		return raw
	}
	return b.ResolveReference(u).String()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func unwrapLumaOGImage(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	host := strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))
	if host != "og.luma.com" {
		return raw
	}
	img := strings.TrimSpace(u.Query().Get("img"))
	if img == "" {
		return raw
	}
	img, _ = url.QueryUnescape(html.UnescapeString(img))
	if img == "" {
		return raw
	}
	return img
}
