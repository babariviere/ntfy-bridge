package bridge

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

type FluxNotification struct {
	InvolvedObject struct {
		Kind            string `json:"kind"`
		Namespace       string `json:"namespace"`
		Name            string `json:"name"`
		UID             string `json:"uid"`
		APIVersion      string `json:"apiVersion"`
		ResourceVersion string `json:"resourceVersion"`
	} `json:"involvedObject"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Reason    string    `json:"reason"`
	Metadata  struct {
		CommitStatus string `json:"commit_status"`
		Revision     string `json:"revision"`
		Summary      string `json:"summary"`
	} `json:"metadata"`
	ReportingController string `json:"reportingController"`
	ReportingInstance   string `json:"reportingInstance"`
}

type FluxHandler struct{}

func NewFluxHandler() FluxHandler {
	return FluxHandler{}
}

func (f FluxHandler) FormatNotification(r io.Reader) (Notification, error) {
	l := slog.With(slog.String("handler", "flux"))
	dec := json.NewDecoder(r)

	var not FluxNotification
	if err := dec.Decode(&not); err != nil {
		l.Error("invalid message format in flux", "error", err)
		return Notification{}, err
	}

	if not.Reason == "ReconciliationSucceeded" {
		// Filter out spammy ReconciliationSucceeded notification
		return Notification{}, nil
	}

	title := fmt.Sprintf("[%s] %s %s/%s.%s", not.Severity, not.Reason,
		strings.ToLower(not.InvolvedObject.Kind), not.InvolvedObject.Namespace, not.InvolvedObject.Name)
	body := not.Message + "\n\n**revision**\n" + not.Metadata.Revision

	l.Debug("flux notification", slog.Group("notification",
		slog.String("title", title),
		slog.String("body", body)))
	return Notification{
		Title:      title,
		Body:       body,
		IsMarkdown: true,
	}, nil
}
