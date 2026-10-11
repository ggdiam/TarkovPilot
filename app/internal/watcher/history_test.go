package watcher

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScanQuestHistoryGroupsProfilesAndModes(t *testing.T) {
	root := t.TempDir()
	oldSession := filepath.Join(root, "log_2025.12.31_23-00-00_1.0.0")
	newSession := filepath.Join(root, "log_2026.01.02_9-00-00_1.0.0")
	mustMkdir(t, oldSession)
	mustMkdir(t, newSession)

	writeTestLog(t, filepath.Join(oldSession, "old application_000.log"),
		"2025-12-31 23:00:01.000|Info|application|CompleteSelectedProfile ProfileId:aaaaaaaaaaaaaaaaaaaaaaaa\n")
	writeTestLog(t, filepath.Join(oldSession, "old backend_000.log"),
		"2025-12-31 23:00:00.000|Info|backend|URL: https://gw-pvp.example/client/game/start\n")
	writeTestLog(t, filepath.Join(oldSession, "old push-notifications_000.log"), testNotification(
		"2025-12-31 23:00:02.000", "12", "111111111111111111111111"))

	writeTestLog(t, filepath.Join(newSession, "new backend_000.log"), strings.Join([]string{
		"2026-01-02 9:00:01.000|Info|backend|URL: https://gw-pve.example/client/game/start",
		"2026-01-02 9:01:00.000|Info|backend|URL: https://gw-pvp-season.example/client/game/start",
	}, "\n")+"\n")
	writeTestLog(t, filepath.Join(newSession, "new application_000.log"), strings.Join([]string{
		"2026-01-02 9:00:02.000|Info|application|CompleteSelectedProfile ProfileId:aaaaaaaaaaaaaaaaaaaaaaaa",
		"2026-01-02 9:01:01.000|Info|application|CompleteSelectedProfile ProfileId:bbbbbbbbbbbbbbbbbbbbbbbb",
	}, "\n")+"\n")
	writeTestLog(t, filepath.Join(newSession, "new push-notifications_000.log"),
		testNotification("2026-01-02 9:00:03.000", "12", "222222222222222222222222")+
			testNotification("2026-01-02 9:01:02.000", "12", "333333333333333333333333")+
			testNotification("2026-01-02 9:01:03.000", "2", "444444444444444444444444"))

	result, err := ScanQuestHistory(root, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Profiles) != 2 {
		t.Fatalf("profiles = %d, want 2", len(result.Profiles))
	}
	if result.Profiles[0].GameMode != "PVE" || result.Profiles[1].GameMode != "SEASON" {
		t.Fatalf("modes = %q, %q", result.Profiles[0].GameMode, result.Profiles[1].GameMode)
	}
	if len(result.Profiles[0].QuestIds) != 1 || result.Profiles[0].QuestIds[0] != "222222222222222222222222" {
		t.Fatalf("unexpected PVE quests: %#v", result.Profiles[0].QuestIds)
	}
	if result.Profiles[0].ProfileKey != hashProfileId("aaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatal("profile key is not the expected hash")
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "aaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatal("raw profile id leaked into scan result")
	}
}

func TestScanQuestHistoryCountsUnassignedCompletion(t *testing.T) {
	root := t.TempDir()
	session := filepath.Join(root, "log_2026.02.01_10-00-00_1.0.0")
	mustMkdir(t, session)
	writeTestLog(t, filepath.Join(session, "test push-notifications_000.log"),
		testNotification("2026-02-01 10:00:00.000", "12", "555555555555555555555555"))

	result, err := ScanQuestHistory(root, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if result.SkippedEvents != 1 || len(result.Profiles) != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestScanQuestHistoryPrestigeResponseBoundary(t *testing.T) {
	for _, test := range []struct {
		name, response string
		wantQuests     int
	}{
		{"empty response", testPrestigeResponse("2026-02-01 10:01:00.000", ""), 1},
		{"successful JSON", testPrestigeResponse("2026-02-01 10:01:00.000", `{"err":0,"data":null}`), 1},
		{"multiline success", testPrestigeResponse("2026-02-01 10:01:00.000", "\n{\n\"err\": 0\n}"), 1},
		{"business error", testPrestigeResponse("2026-02-01 10:01:00.000", `{"err":228,"errmsg":"unavailable"}`), 2},
		{"multiline error", testPrestigeResponse("2026-02-01 10:01:00.000", "\n{\n\"err\": 228\n}"), 2},
		{"error field", testPrestigeResponse("2026-02-01 10:01:00.000", `{"error":"failed"}`), 2},
		{"malformed body", testPrestigeResponse("2026-02-01 10:01:00.000", "invalid JSON"), 2},
		{"request only", "2026-02-01 10:01:00.000|Info|backend|---> Request HTTPS, id [1]: URL: https://gw-pve.example/client/prestige/obtain.\n", 2},
		{"prestige list", strings.ReplaceAll(testPrestigeResponse("2026-02-01 10:01:00.000", ""), "/prestige/obtain", "/prestige/list"), 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			session := filepath.Join(root, "log_2026.02.01_10-00-00_1.0.0")
			writeHistorySession(t, session, "PVE", "aaaaaaaaaaaaaaaaaaaaaaaa", "2026-02-01 10:00:00.000",
				test.response,
				testNotification("2026-02-01 10:00:30.000", "12", "111111111111111111111111")+
					testNotification("2026-02-01 10:02:00.000", "12", "222222222222222222222222"))
			result, err := ScanQuestHistory(root, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Profiles) != 1 || len(result.Profiles[0].QuestIds) != test.wantQuests {
				t.Fatalf("unexpected scan: %#v", result)
			}
			if test.wantQuests == 1 && result.Profiles[0].QuestIds[0] != "222222222222222222222222" {
				t.Fatalf("old quest survived reset: %#v", result.Profiles[0])
			}
			if test.wantQuests == 1 && result.Profiles[0].FirstEventAt.Format("15:04:05") != "10:02:00" {
				t.Fatalf("old event date survived reset: %#v", result.Profiles[0])
			}
		})
	}
}

func TestScanQuestHistoryLatestPrestigeAcrossUnorderedSessions(t *testing.T) {
	root := t.TempDir()
	profile := "aaaaaaaaaaaaaaaaaaaaaaaa"
	// ReadDir orders 0, 19, 3; neither clearing during traversal nor using the last seen reset is safe.
	writeHistorySession(t, filepath.Join(root, "log_2026.02.01_0-00-00_1.0.0"), "PVE", profile,
		"2026-02-01 00:00:00.000", testPrestigeResponse("2026-02-01 00:01:00.000", ""),
		testNotification("2026-02-01 00:00:30.000", "12", "111111111111111111111111")+
			testNotification("2026-02-01 00:02:00.000", "12", "222222222222222222222222"))
	writeHistorySession(t, filepath.Join(root, "log_2026.02.01_19-00-00_1.0.0"), "PVE", profile,
		"2026-02-01 19:00:00.000", testPrestigeResponse("2026-02-01 19:01:00.000", ""),
		testNotification("2026-02-01 19:01:00.000", "12", "333333333333333333333333")+
			testNotification("2026-02-01 19:02:00.000", "12", "111111111111111111111111"))
	writeHistorySession(t, filepath.Join(root, "log_2026.02.01_3-00-00_1.0.0"), "PVE", profile,
		"2026-02-01 03:00:00.000", testPrestigeResponse("2026-02-01 03:01:00.000", ""),
		testNotification("2026-02-01 03:02:00.000", "12", "444444444444444444444444"))
	// A reset must not remove another profile or mode, even with the same quest ID.
	writeHistorySession(t, filepath.Join(root, "log_2026.02.01_1-00-00_1.0.0"), "PVE", "bbbbbbbbbbbbbbbbbbbbbbbb",
		"2026-02-01 01:00:00.000", "", testNotification("2026-02-01 01:02:00.000", "12", "555555555555555555555555"))
	writeHistorySession(t, filepath.Join(root, "log_2026.02.01_2-00-00_1.0.0"), "SEASON", profile,
		"2026-02-01 02:00:00.000", "", testNotification("2026-02-01 02:02:00.000", "12", "666666666666666666666666"))
	result, err := ScanQuestHistory(root, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Profiles) != 3 || result.SkippedEvents != 0 {
		t.Fatalf("unexpected scan: %#v", result)
	}
	for _, got := range result.Profiles {
		want := "111111111111111111111111"
		if got.GameMode == "SEASON" {
			want = "666666666666666666666666"
		} else if got.ProfileKey == hashProfileId("bbbbbbbbbbbbbbbbbbbbbbbb") {
			want = "555555555555555555555555"
		}
		if len(got.QuestIds) != 1 || got.QuestIds[0] != want {
			t.Fatalf("unexpected quests: %#v", got)
		}
	}
}

func TestScanQuestHistoryPrestigeWithoutNewCompletions(t *testing.T) {
	root := t.TempDir()
	writeHistorySession(t, filepath.Join(root, "log_2026.02.01_10-00-00_1.0.0"), "PVE", "aaaaaaaaaaaaaaaaaaaaaaaa",
		"2026-02-01 10:00:00.000", testPrestigeResponse("2026-02-01 10:01:00.000", ""),
		testNotification("2026-02-01 10:00:30.000", "12", "111111111111111111111111"))
	result, err := ScanQuestHistory(root, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Profiles) != 0 || result.SkippedEvents != 0 {
		t.Fatalf("old history returned: %#v", result)
	}
}

func testPrestigeResponse(stamp, body string) string {
	return stamp + "|Info|backend|<--- Response HTTPS, id [1]: URL: https://gw-pve.example/client/prestige/obtain, DownloadSeconds: 0.7, ParseSeconds: 0.008, SumSeconds: 0.708, responseText: " + body + "\n"
}

func writeHistorySession(t *testing.T, folder, mode, profile, stamp, backend, notifications string) {
	t.Helper()
	mustMkdir(t, folder)
	gateway := strings.ToLower(mode)
	if mode == "SEASON" {
		gateway = "pvp-season"
	}
	writeTestLog(t, filepath.Join(folder, "test backend_000.log"), stamp+"|Info|backend|URL: https://gw-"+gateway+".example/client/game/start\n"+backend)
	writeTestLog(t, filepath.Join(folder, "test application_000.log"), stamp+"|Info|application|CompleteSelectedProfile ProfileId:"+profile+"\n")
	writeTestLog(t, filepath.Join(folder, "test push-notifications_000.log"), notifications)
}

func TestLiveQuestIncludesAnonymizedProfileRouting(t *testing.T) {
	w := NewLogsWatcher()
	var questId, status, profileKey, gameMode string
	w.OnQuest = func(gotQuestId, gotStatus, gotProfileKey, gotGameMode string) {
		questId, status, profileKey, gameMode = gotQuestId, gotStatus, gotProfileKey, gotGameMode
	}

	w.processLines([]string{
		"2026-02-01 10:00:00.000|Info|backend|URL: https://gw-pve.example/client/game/start",
		"2026-02-01 10:00:01.000|Info|application|CompleteSelectedProfile ProfileId:aaaaaaaaaaaaaaaaaaaaaaaa",
		"2026-02-01 10:00:02.000|Info|push-notifications|Got notification | ChatMessageReceived",
		`{"message":{"type":12,"templateId":"555555555555555555555555 successMessageText"}}`,
	})

	if questId != "555555555555555555555555" || status != "12" || gameMode != "PVE" {
		t.Fatalf("unexpected quest routing: quest=%q status=%q mode=%q", questId, status, gameMode)
	}
	if profileKey != hashProfileId("aaaaaaaaaaaaaaaaaaaaaaaa") || profileKey == "aaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatal("live quest did not use the anonymized profile key")
	}
}

func TestScanQuestHistoryFromEnvironment(t *testing.T) {
	root := os.Getenv("EFT_LOGS_TEST_PATH")
	if root == "" {
		t.Skip("EFT_LOGS_TEST_PATH is not set")
	}

	result, err := ScanQuestHistory(root, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Profiles) == 0 {
		t.Fatal("no anonymized profiles found in configured logs")
	}
	quests := 0
	modes := map[string]int{}
	profileModes := map[string]string{}
	crossModeProfiles := 0
	for _, profile := range result.Profiles {
		quests += len(profile.QuestIds)
		modes[profile.GameMode]++
		if previous := profileModes[profile.ProfileKey]; previous != "" && previous != profile.GameMode {
			crossModeProfiles++
		}
		profileModes[profile.ProfileKey] = profile.GameMode
		t.Logf("mode=%s quests=%d first=%s last=%s", profile.GameMode, len(profile.QuestIds),
			profile.FirstEventAt.Format("2006-01-02"), profile.LastEventAt.Format("2006-01-02"))
	}
	t.Logf("profiles=%d quests=%d skipped=%d crossModeProfiles=%d modes=%v",
		len(result.Profiles), quests, result.SkippedEvents, crossModeProfiles, modes)
}

func testNotification(stamp, status, questId string) string {
	return stamp + "|Info|push-notifications|Got notification | ChatMessageReceived\n" +
		`{"message":{"type":` + status + `,"templateId":"` + questId + ` successMessageText"}}` + "\n"
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeTestLog(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
