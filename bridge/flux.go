package bridge

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type fluxInvolvedObject struct {
	Kind            string `json:"kind"`
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	UID             string `json:"uid"`
	APIVersion      string `json:"apiVersion"`
	ResourceVersion string `json:"resourceVersion"`
}

func (f fluxInvolvedObject) String() string {
	return strings.ToLower(f.Kind) + "/" + f.Namespace + "." + f.Name
}

type FluxNotification struct {
	InvolvedObject fluxInvolvedObject `json:"involvedObject"`
	Severity       string             `json:"severity"`
	Timestamp      time.Time          `json:"timestamp"`
	Message        string             `json:"message"`
	Reason         string             `json:"reason"`
	Metadata       struct {
		CommitStatus string `json:"commit_status"`
		Revision     string `json:"revision"`
		Summary      string `json:"summary"`
	} `json:"metadata"`
	ReportingController string `json:"reportingController"`
	ReportingInstance   string `json:"reportingInstance"`
}

type FluxHandler struct {
	// Register all modifications of reconciliations
	reconciliations map[string]bool
}

func NewFluxHandler() FluxHandler {
	return FluxHandler{
		reconciliations: make(map[string]bool),
	}
}

func (f FluxHandler) ProduceNotifications(r *http.Request) ([]Notification, error) {
	l := slog.With(slog.String("handler", "flux"))
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var not FluxNotification
	if err := dec.Decode(&not); err != nil {
		l.Error("invalid message format", "error", err)
		return nil, err
	}

	obj := not.InvolvedObject.String()
	if not.Reason == "ReconciliationSucceeded" {
		if ok := f.reconciliations[obj]; !ok {
			// Filter out spammy ReconciliationSucceeded notification
			return nil, errSkipNotification
		}

		// we will print the object so skip it next time it spam
		f.reconciliations[obj] = false
	} else {
		// object has been modified, we can print it next time
		f.reconciliations[obj] = true
	}

	title := fmt.Sprintf("[%s] %s %s", not.Severity, not.Reason, obj)
	body := not.Message + "\n\n**revision**\n" + not.Metadata.Revision

	l.Debug("flux notification", slog.Group("notification",
		slog.String("title", title),
		slog.String("body", body)))
	return []Notification{{
		Title:      title,
		Body:       body,
		IsMarkdown: true,
	}}, nil
}
