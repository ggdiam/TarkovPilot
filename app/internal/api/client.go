// Package api — sending webhooks to the TM website.
// Single endpoint for the app and the ps1 script: /api/be/pilot/script/webhook.
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// ErrBadKey — the server replied that hookId was not found (connection key is invalid)
var ErrBadKey = errors.New("bad key")

type Client struct {
	// getters instead of copies: hookId and host are edited in the UI on the fly
	HookId func() string
	Host   func() string

	// last send succeeded (for the connection status in the UI)
	serverOK atomic.Bool
	// whether at least one send attempt happened (before it the UI shows "checking", not an error)
	checked atomic.Bool
	// last status webhook replied bad_key — the key was rejected by the server
	badKey atomic.Bool

	http *http.Client
}

func NewClient(hookId, host func() string) *Client {
	return &Client{
		HookId: hookId,
		Host:   host,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) ServerOK() bool      { return c.serverOK.Load() }
func (c *Client) ServerChecked() bool { return c.checked.Load() }
func (c *Client) BadKey() bool        { return c.badKey.Load() }

// ResetCheck — reset to the "checking" state (after the key or host changed)
func (c *Client) ResetCheck() {
	c.checked.Store(false)
	c.badKey.Store(false)
}

// AppStatus — the app block of the status event; fields mirror PilotAppStatus on the website
type AppStatus struct {
	Version             string `json:"version"`
	GameFolder          string `json:"gameFolder"`
	ScreenshotsFolder   string `json:"screenshotsFolder"`
	GameFound           bool   `json:"gameFound"`
	LogsWatching        bool   `json:"logsWatching"`
	ScreenshotsWatching bool   `json:"screenshotsWatching"`
}

type statusResponse struct {
	Ok            bool   `json:"ok"`
	Error         string `json:"error"`
	LatestVersion string `json:"latestVersion"`
}

func (c *Client) url() string {
	return fmt.Sprintf("https://%s/api/be/pilot/script/webhook", c.Host())
}

// post sends a webhook; 2 attempts — events are few and the network can blink
func (c *Client) post(payload map[string]any) ([]byte, error) {
	payload["hookId"] = c.HookId()

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(2 * time.Second)
		}

		resp, err := c.http.Post(c.url(), "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = err
			continue
		}

		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.serverOK.Store(true)
			c.checked.Store(true)
			return buf.Bytes(), nil
		}
		lastErr = fmt.Errorf("http %d", resp.StatusCode)
	}

	c.serverOK.Store(false)
	c.checked.Store(true)
	return nil, lastErr
}

// SendScreenshot — player position: the website parses coordinates from the file name
func (c *Client) SendScreenshot(filename string) error {
	_, err := c.post(map[string]any{"event": "screenshot", "filename": filename})
	return err
}

// SendMap — location change (raw name from the log, e.g. 'bigmap')
func (c *Client) SendMap(mapName string) error {
	_, err := c.post(map[string]any{"event": "map", "map": mapName})
	return err
}

// SendQuest — quest event from the push-notifications log
func (c *Client) SendQuest(questId, status string) error {
	_, err := c.post(map[string]any{"event": "quest", "questId": questId, "status": status})
	return err
}

// SendStatus — heartbeat with app statuses; returns the latest version from the website.
// A bad_key response (key not found on the server) is returned as ErrBadKey
func (c *Client) SendStatus(st AppStatus) (latestVersion string, err error) {
	b, err := c.post(map[string]any{"event": "status", "app": st})
	if err != nil {
		return "", err
	}
	var resp statusResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return "", nil // an older server could reply with a plain 'ok' string — not an error
	}
	if !resp.Ok && resp.Error == "bad_key" {
		c.badKey.Store(true)
		return resp.LatestVersion, ErrBadKey
	}
	c.badKey.Store(false)
	return resp.LatestVersion, nil
}

// FetchLatestVersion — version from the website without a key (GET /api/be/pilot/version, plain text).
// A separate channel from the heartbeat: works without hookId and doesn't touch connection statuses
func (c *Client) FetchLatestVersion() (string, error) {
	url := fmt.Sprintf("https://%s/api/be/pilot/version", c.Host())
	resp, err := c.http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return "", err
	}
	return strings.Trim(strings.TrimSpace(string(b)), `"`), nil
}
