package api

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

type PlayerEvent struct {
	Time   string `json:"time"`
	Player string `json:"player"`
	Event  string `json:"event"` // "join", "leave", "chat", "death"
	Detail string `json:"detail,omitempty"`
}

var (
	// [00:12:34]: [Join Announcement] PlayerName
	joinRegex = regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\].*\[Join Announcement\]\s+(.+)`)
	// [00:12:34]: [Leave Announcement] PlayerName
	leaveRegex = regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\].*\[Leave Announcement\]\s+(.+)`)
	// [00:12:34]: [Say] (PlayerName) message
	chatRegex = regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\].*\[Say\]\s+\(([^)]+)\)\s+(.+)`)
	// [00:12:34]: [Death] PlayerName was killed by ...
	deathRegex = regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\].*\[Death\]\s+(.+)`)
)

func (h *Handler) GetPlayerActivity(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "clusterID")
	cluster, err := h.store.GetCluster(clusterID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	logPath := filepath.Join(h.dataDir, "clusters", cluster.DirName, "Master", "server_log.txt")
	chatLogPath := filepath.Join(h.dataDir, "clusters", cluster.DirName, "Master", "server_chat_log.txt")

	var events []PlayerEvent

	// Parse server_log.txt for join/leave/death
	events = append(events, parseLogFile(logPath)...)
	// Parse chat log
	events = append(events, parseChatLog(chatLogPath)...)

	// Return last 100 events (newest first)
	if len(events) > 100 {
		events = events[len(events)-100:]
	}
	// Reverse for newest first
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	writeJSON(w, http.StatusOK, events)
}

func parseLogFile(path string) []PlayerEvent {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var events []PlayerEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := joinRegex.FindStringSubmatch(line); m != nil {
			events = append(events, PlayerEvent{Time: m[1], Player: strings.TrimSpace(m[2]), Event: "join"})
		} else if m := leaveRegex.FindStringSubmatch(line); m != nil {
			events = append(events, PlayerEvent{Time: m[1], Player: strings.TrimSpace(m[2]), Event: "leave"})
		} else if m := deathRegex.FindStringSubmatch(line); m != nil {
			events = append(events, PlayerEvent{Time: m[1], Player: strings.TrimSpace(m[2]), Event: "death"})
		}
	}
	return events
}

func parseChatLog(path string) []PlayerEvent {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var events []PlayerEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if m := chatRegex.FindStringSubmatch(line); m != nil {
			events = append(events, PlayerEvent{Time: m[1], Player: m[2], Event: "chat", Detail: m[3]})
		}
	}
	return events
}
