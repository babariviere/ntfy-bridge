package bridge

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

type DiscordMessage struct {
	Content string `json:"content"`
	Embeds  []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Footer      struct {
			Text string `json:"text"`
		} `json:"footer"`
		Author struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"author"`
	} `json:"embeds"`
}

type DiscordEmbedHandler struct{}

func NewDiscordEmbedHandler() DiscordEmbedHandler {
	return DiscordEmbedHandler{}
}

func (d DiscordEmbedHandler) ProduceNotifications(r *http.Request) ([]Notification, error) {
	l := slog.With(slog.String("handler", "discord_embed"))

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var not DiscordMessage
	if err := dec.Decode(&not); err != nil {
		l.Error("invalid message format", "error", err)
		return nil, err
	}

	notifications := make([]Notification, len(not.Embeds))
	for i, embed := range not.Embeds {
		not := notifications[i]
		not.Title = embed.Title
		not.IsMarkdown = true
		if embed.URL != "" {
			not.Actions = []NotificationAction{NewViewAction("Open in Browser", embed.URL)}
		}

		var body strings.Builder
		body.WriteString(embed.Description)

		if embed.Author.Name != "" {
			body.WriteString("\n\n**Author**\n")
			body.WriteString(embed.Author.Name)
			if embed.Author.URL != "" {
				body.WriteString(" (" + embed.Author.URL + ")")
			}
		}

		if embed.Footer.Text != "" {
			body.WriteString("\n\n" + embed.Footer.Text)
		}

		not.Body = body.String()

		notifications[i] = not
	}

	return notifications, nil
}
