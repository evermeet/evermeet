package email

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"net/mail"
)

func findSendmail() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	candidates := []string{
		"/usr/sbin/sendmail",
		"/usr/lib/sendmail",
		"/sbin/sendmail",
	}
	for _, p := range candidates {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	if p, err := exec.LookPath("sendmail"); err == nil {
		return p
	}
	return ""
}

func deliverSendmail(path, from string, msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	args := []string{"-t", "-i"}
	if addr := strings.TrimSpace(from); addr != "" {
		fromAddr := addr
		if a, err := mail.ParseAddress(addr); err == nil {
			fromAddr = a.Address
		}
		args = append([]string{"-f", fromAddr}, args...)
	}

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Stdin = bytes.NewReader(msg)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		out := strings.TrimSpace(stderr.String())
		if out != "" {
			return fmt.Errorf("sendmail: %w (%s)", err, out)
		}
		return fmt.Errorf("sendmail: %w", err)
	}
	return nil
}
