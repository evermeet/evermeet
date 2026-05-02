package node

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Default HTTPS endpoints that return the caller's public IPv4 as plain text.
// Docker does not expose the host's routable address inside the container; outbound
// lookup matches what remote peers would use when dialling your published port.
var defaultPublicIPURLs = []string{
	"https://api.ipify.org",
	"https://ifconfig.me/ip",
}

// AppendPublicAnnounceAddr appends /ip4/<public-ipv4>/tcp/<listenPort> to existing
// announce multiaddrs if that entry is not already present.
func AppendPublicAnnounceAddr(ctx context.Context, log *log.Logger, listenPort int, existing []string) ([]string, error) {
	ip, err := discoverPublicIPv4(ctx)
	if err != nil {
		return existing, err
	}
	suffix := fmt.Sprintf("/ip4/%s/tcp/%d", ip, listenPort)
	for _, e := range existing {
		if strings.TrimSpace(e) == suffix {
			return existing, nil
		}
	}
	out := append(append([]string(nil), existing...), suffix)
	if log != nil {
		log.Printf("p2p: auto announce %s (public IPv4 via HTTPS)", suffix)
	}
	return out, nil
}

func discoverPublicIPv4(ctx context.Context) (string, error) {
	urls := defaultPublicIPURLs
	if u := strings.TrimSpace(os.Getenv("EVERMEET_PUBLIC_IP_URL")); u != "" {
		urls = []string{u}
	}
	client := &http.Client{Timeout: 6 * time.Second}
	var lastErr error
	for _, u := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			lastErr = err
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("%s: HTTP %d", u, resp.StatusCode)
			continue
		}
		ipStr := strings.TrimSpace(string(body))
		if pip := net.ParseIP(ipStr); pip != nil && pip.To4() != nil {
			return pip.String(), nil
		}
		lastErr = fmt.Errorf("%s: expected IPv4, got %q", u, ipStr)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no public IP lookup URLs succeeded")
	}
	return "", lastErr
}
