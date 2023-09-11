package bridge

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type AlertmanagerEvent struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   []struct {
		Status       string            `json:"status"`
		Labels       map[string]string `json:"labels"`
		Annotations  map[string]string `json:"annotations"`
		StartsAt     time.Time         `json:"startsAt"`
		EndsAt       time.Time         `json:"endsAt"`
		GeneratorURL string            `json:"generatorURL"`
		Fingerprint  string            `json:"fingerprint"`
	} `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

type AlertmanagerHandler struct{}

func NewAlertmanagerHandler() AlertmanagerHandler {
	return AlertmanagerHandler{}
}

func (d AlertmanagerHandler) ProduceNotifications(r *http.Request) ([]Notification, error) {
	l := slog.With(slog.String("handler", "alertmanager"))

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var event AlertmanagerEvent
	if err := dec.Decode(&event); err != nil {
		l.Error("invalid message format", "error", err)
		return nil, err
	}

	notifications := make([]Notification, 0, len(event.Alerts))
	for _, alert := range event.Alerts {
		if alert.Annotations["summary"] == "" {
			continue
		}

		var not Notification
		not.Title = "[" + strings.ToUpper(event.Status) + "] " + alert.Annotations["summary"]
		not.Body = alert.Annotations["description"]
		if runbook := alert.Annotations["runbook_url"]; runbook != "" {
			not.Actions = append(not.Actions, NewViewAction("Runbook", runbook))
		}
		if event.Status == "resolved" {
			not.Tags = []string{"resolved"}
		}

		notifications = append(notifications, not)
	}

	return notifications, nil
}
