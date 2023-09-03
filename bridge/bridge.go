package bridge

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

var (
	errSkipNotification = errors.New("notification skipped")
)

type Handler interface {
	ProduceNotifications(r *http.Request) ([]Notification, error)
}

type NotificationError struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Notification struct {
	Title      string
	Body       string
	Priority   int
	Tags       []string
	IsMarkdown bool

	topic string
	auth  Auth
}

func (n Notification) IsEmpty() bool {
	return n.Title == "" && n.Body == ""
}

func (n Notification) Send(base string) error {
	req, err := http.NewRequest("POST", base+"/"+n.topic, strings.NewReader(n.Body))
	if err != nil {
		return err
	}

	if n.IsMarkdown {
		req.Header.Set("Content-Type", "text/markdown")
	} else {
		req.Header.Set("Content-Type", "text/plain")
	}

	if n.Title != "" {
		req.Header.Set("Title", n.Title)
	}

	if n.Priority > 0 {
		req.Header.Set("Priority", strconv.Itoa(n.Priority))
	}

	if len(n.Tags) > 0 {
		req.Header.Set("Tags", strings.Join(n.Tags, ","))
	}

	if n.auth.Username != "" {
		req.Header.Set("Authorization", n.auth.basic())
	}

	if n.auth.AccessToken != "" {
		req.Header.Set("Authorization", n.auth.bearer())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var e NotificationError
		dec := json.NewDecoder(resp.Body)
		dec.Decode(&e)

		return errors.New(e.Error)
	}

	return nil
}

type Auth struct {
	Username string
	Password string

	AccessToken string
}

func (a Auth) IsEmpty() bool {
	return a.Username == "" && a.AccessToken == ""
}

func (a Auth) basic() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(a.Username+":"+a.Password))
}

func (a Auth) bearer() string {
	return "Bearer " + a.AccessToken
}

type Bridge struct {
	baseURL string
	topic   string
	h       Handler
	auth    Auth
}

func NewBridge(baseURL, topic string, handler Handler) Bridge {
	return Bridge{
		baseURL: baseURL,
		topic:   topic,
		h:       handler,
	}
}

func (b *Bridge) WithAuth(auth Auth) {
	b.auth = auth
}

func (b Bridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	nots, err := b.h.ProduceNotifications(r)

	if errors.Is(err, errSkipNotification) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err != nil {
		slog.Error("failed to format notification")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, not := range nots {
		not.topic = b.topic
		not.auth = b.auth
		if err = not.Send(b.baseURL); err != nil {
			slog.Error("unable to send notification", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	slog.Debug("notifications sent with success", "sent", len(nots))

	w.WriteHeader(http.StatusNoContent)
}
