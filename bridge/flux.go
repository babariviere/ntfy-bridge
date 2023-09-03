package bridge

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// TODO: use skaffold to see why the revision is not present
type FluxNotification struct {
	InvolvedObject struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Name       string `json:"name"`
		Namespace  string `json:"namespace"`
		UID        string `json:"uid"`
	} `json:"involvedObject"`
	Metadata struct {
		KustomizeToolkitFluxcdIoRevision string `json:"kustomize.toolkit.fluxcd.io/revision"`
	} `json:"metadata"`
	Severity            string    `json:"severity"`
	Reason              string    `json:"reason"`
	Message             string    `json:"message"`
	ReportingController string    `json:"reportingController"`
	ReportingInstance   string    `json:"reportingInstance"`
	Timestamp           time.Time `json:"timestamp"`
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
	body := not.Message + "\n\n**revision**\n" + not.Metadata.KustomizeToolkitFluxcdIoRevision

	l.Debug("flux notification", slog.Group("notification",
		slog.String("title", title),
		slog.String("body", body)))
	return Notification{
		Title:      title,
		Body:       body,
		IsMarkdown: true,
	}, nil
}
