package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var discordWebhookURL string

func InitDiscord(webhookURL string) {
	discordWebhookURL = webhookURL
	if webhookURL != "" {
		log.Printf("Discord notifications enabled")
	}
}

func NotifyDiscord(title, description string, color int) {
	if discordWebhookURL == "" {
		return
	}

	go func() {
		payload := map[string]any{
			"embeds": []map[string]any{
				{
					"title":       title,
					"description": description,
					"color":       color,
					"timestamp":   time.Now().Format(time.RFC3339),
					"footer": map[string]string{
						"text": "DST DS Panel",
					},
				},
			},
		}

		data, _ := json.Marshal(payload)
		resp, err := http.Post(discordWebhookURL, "application/json", bytes.NewReader(data))
		if err != nil {
			log.Printf("Discord notification failed: %v", err)
			return
		}
		resp.Body.Close()
	}()
}

// Color constants for Discord embeds
const (
	ColorGreen  = 0x22c55e // success
	ColorRed    = 0xef4444 // error
	ColorYellow = 0xeab308 // warning
	ColorBlue   = 0x3b82f6 // info
)

func NotifyServerStarted(clusterName string) {
	NotifyDiscord(
		"Server Started",
		fmt.Sprintf("Cluster **%s** is now running.", clusterName),
		ColorGreen,
	)
}

func NotifyServerStopped(clusterName string) {
	NotifyDiscord(
		"Server Stopped",
		fmt.Sprintf("Cluster **%s** has been stopped.", clusterName),
		ColorRed,
	)
}

func NotifyServerError(clusterName, errMsg string) {
	NotifyDiscord(
		"Server Error",
		fmt.Sprintf("Cluster **%s** encountered an error:\n```%s```", clusterName, errMsg),
		ColorYellow,
	)
}
