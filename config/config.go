package config

import (
	"errors"
	"fmt"

	"git.sr.ht/~emersion/go-scfg"
)

type HandlerType int

const (
	HandlerNone HandlerType = iota
	HandlerFlux
)

type LogFormat int

const (
	LogFormatText LogFormat = iota
	LogFormatJson
)

type Config struct {
	HTTPAddress string
	LogLevel    string
	LogFormat   LogFormat

	Ntfy     ntfy
	Handlers map[string]handler
}

type ntfy struct {
	Server       string
	DefaultTopic string
}

func (n *ntfy) readConfig(cfg scfg.Block) error {
	if err := readString(cfg, "server", &n.Server); err != nil {
		return err
	}

	return readString(cfg, "default-topic", &n.DefaultTopic)
}

type handler struct {
	Type  HandlerType
	Topic string
}

func (h *handler) readConfig(cfg scfg.Block) error {
	var err error
	if h.Type, err = readHandlerType(cfg.Get("type")); err != nil {
		return err
	}

	return readStringRequired(cfg, "topic", &h.Topic)
}

func ReadConfig(path string) (Config, error) {
	cfg, err := scfg.Load(path)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		HTTPAddress: "127.0.0.1:8080",
		LogLevel:    "info",
		LogFormat:   LogFormatText,

		Ntfy: ntfy{
			Server: "https://ntfy.sh",
		},
		Handlers: make(map[string]handler),
	}

	if err := readString(cfg, "http-address", &config.HTTPAddress); err != nil {
		return config, err
	}

	if err := readString(cfg, "log-level", &config.LogLevel); err != nil {
		return config, err
	}

	if err := readLogFormat(cfg, "log-format", &config.LogFormat); err != nil {
		return config, err
	}

	d := cfg.Get("ntfy")
	if d != nil {
		if err := config.Ntfy.readConfig(d.Children); err != nil {
			return config, err
		}
	}

	ds := cfg.GetAll("handler")
	for _, d := range ds {
		var key string
		if err := d.ParseParams(&key); err != nil {
			return config, err
		}

		var h handler
		if err := h.readConfig(d.Children); err != nil {
			return config, err
		}

		config.Handlers[key] = h
	}

	for _, block := range cfg {
		fmt.Printf("%+v\n", block)

	}
	return config, nil
}

func readHandlerType(d *scfg.Directive) (HandlerType, error) {
	if d == nil {
		return HandlerNone, errors.New("handler.type is missing")
	}

	var ty string

	if err := d.ParseParams(&ty); err != nil {
		return HandlerNone, err
	}

	switch ty {
	case "flux":
		return HandlerFlux, nil
	default:
		return HandlerNone, fmt.Errorf("invalid handler type %q", ty)
	}
}

func readLogFormat(cfg scfg.Block, key string, val *LogFormat) error {
	d := cfg.Get(key)
	if d == nil {
		return nil
	}

	var format string
	if err := d.ParseParams(&format); err != nil {
		return err
	}

	switch format {
	case "text":
		*val = LogFormatText
	case "json":
		*val = LogFormatJson
	default:
		return fmt.Errorf("invalid log format %q", format)
	}
	return nil
}

func readString(cfg scfg.Block, key string, val *string) error {
	d := cfg.Get(key)
	if d != nil {
		return d.ParseParams(val)
	}
	return nil
}

func readStringRequired(cfg scfg.Block, key string, val *string) error {
	d := cfg.Get(key)
	if d != nil {
		return d.ParseParams(val)
	}
	return fmt.Errorf("missing key %q", key)
}
