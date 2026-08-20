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
	// whether the connected website account has an active PRO subscription
	pro atomic.Bool

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
func (c *Client) Pro() bool           { return c.pro.Load() }

// ResetCheck — reset to the "checking" state (after the key or host changed)
func (c *Client) ResetCheck() {
	c.checked.Store(false)
	c.badKey.Store(false)
	c.pro.Store(false)
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
	Pro           bool   `json:"pro"`
}

// QuestSyncCharacter is a website character available as an import target.
type QuestSyncCharacter struct {
	UID      string `json:"uid"`
	CharName string `json:"charName"`
	GameMode string `json:"gameMode"`
	Fraction string `json:"fraction"`
	Level    int    `json:"level"`
}

// QuestSyncProfile is an anonymized EFT profile found in local logs.
type QuestSyncProfile struct {
	ProfileKey    string `json:"profileKey"`
	GameMode      string `json:"gameMode"`
	FoundCount    int    `json:"foundCount"`
	MatchedCount  int    `json:"matchedCount"`
	IgnoredCount  int    `json:"ignoredCount"`
	FirstEventAt  string `json:"firstEventAt"`
	LastEventAt   string `json:"lastEventAt"`
	LinkedCharUID string `json:"linkedCharUid"`
}

// QuestImportResult describes one completed import operation.
type QuestImportResult struct {
	CharUID       string `json:"charUid"`
	ImportedCount int    `json:"importedCount"`
	AlreadyDone   int    `json:"alreadyDone"`
	IgnoredCount  int    `json:"ignoredCount"`
}

// QuestSyncResponse is shared by scan, import, and create-and-import requests.
type QuestSyncResponse struct {
	Ok            bool                 `json:"ok"`
	Error         string               `json:"error"`
	Profiles      []QuestSyncProfile   `json:"profiles"`
	Characters    []QuestSyncCharacter `json:"characters"`
	SkippedEvents int                  `json:"skippedEvents"`
	SourceSince   string               `json:"sourceSince"`
	UpdatedAt     string               `json:"updatedAt"`
	ImportResult  *QuestImportResult   `json:"importResult"`
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

// SendQuest — quest event from the push-notifications log with an anonymized profile key.
func (c *Client) SendQuest(questId, status, profileKey, gameMode string) error {
	payload := map[string]any{
		"event":          "quest",
		"questId":        questId,
		"status":         status,
		"profileRouting": true,
	}
	if profileKey != "" {
		payload["profileKey"] = profileKey
	}
	if gameMode != "" {
		payload["gameMode"] = gameMode
	}
	_, err := c.post(payload)
	return err
}

// SendQuestScan uploads an anonymized snapshot and returns import targets from the website.
func (c *Client) SendQuestScan(profiles any, skippedEvents int) (QuestSyncResponse, error) {
	return c.questSyncPost(map[string]any{
		"event":         "quest_scan",
		"profiles":      profiles,
		"skippedEvents": skippedEvents,
	})
}

// ImportQuestProfile imports one scanned profile into an existing website character.
func (c *Client) ImportQuestProfile(profileKey, charUID string) (QuestSyncResponse, error) {
	return c.questSyncPost(map[string]any{
		"event":      "quest_import",
		"profileKey": profileKey,
		"charUid":    charUID,
	})
}

// CreateQuestCharacterAndImport creates a website character and imports the profile into it.
func (c *Client) CreateQuestCharacterAndImport(profileKey, fraction string) (QuestSyncResponse, error) {
	return c.questSyncPost(map[string]any{
		"event":      "quest_create_import",
		"profileKey": profileKey,
		"fraction":   fraction,
	})
}

func (c *Client) questSyncPost(payload map[string]any) (QuestSyncResponse, error) {
	b, err := c.post(payload)
	if err != nil {
		return QuestSyncResponse{}, err
	}
	var resp QuestSyncResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return QuestSyncResponse{}, err
	}
	if !resp.Ok && resp.Error == "bad_key" {
		c.badKey.Store(true)
		c.pro.Store(false)
	} else if !resp.Ok && resp.Error == "pro_required" {
		c.pro.Store(false)
	}
	return resp, nil
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
		c.pro.Store(false)
		return resp.LatestVersion, ErrBadKey
	}
	c.badKey.Store(false)
	c.pro.Store(resp.Pro)
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
