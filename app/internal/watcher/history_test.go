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
