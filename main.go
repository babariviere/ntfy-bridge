package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"forge.babariviere.com/babariviere/ntfy-bridge/bridge"
	"forge.babariviere.com/babariviere/ntfy-bridge/config"
)

// TODO: allow to pass config path
func readConfig() (config.Config, error) {
	paths := []string{
		"config.scfg",
		// TODO: properly handle XDG_CONFIG
		os.Getenv("HOME") + "/.config/ntfy-bridge/config.scfg",
		os.Getenv("HOME") + "/.ntfy-bridge/config.scfg",
		"/etc/ntfy-bridge/config.scfg",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			continue
		}

		return config.ReadConfig(path)
	}

	return config.Config{}, errors.New("no configuration file found")
}

func main() {
	cfg, err := readConfig()
	if err != nil {
		slog.Error("failed to read config", "error", err)
		os.Exit(2)
	}

	defaultLevel := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		defaultLevel = slog.LevelDebug
	case "warn":
		defaultLevel = slog.LevelWarn
	case "error":
		defaultLevel = slog.LevelError
	}
	lopts := slog.HandlerOptions{
		Level: defaultLevel,
	}

	switch cfg.LogFormat {
	case config.LogFormatText:
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &lopts)))
	case config.LogFormatJson:
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &lopts)))
	}

	slog.Info("Successfully read config")

	var auth bridge.Auth
	if cfg.Ntfy.AccessToken != "" {
		auth.AccessToken = cfg.Ntfy.AccessToken
	} else if cfg.Ntfy.Username != "" {
		auth.Username = cfg.Ntfy.Username
		auth.Password = cfg.Ntfy.Password
	}

	for route, handler := range cfg.Handlers {
		var h bridge.Handler
		switch handler.Type {
		case config.HandlerFlux:
			h = bridge.NewFluxHandler()
		case config.HandlerDiscordEmbed:
			h = bridge.NewDiscordEmbedHandler()
		}

		slog.Debug("Registering bridge", "route", route, "handler", handler.Type)
		topic := handler.Topic
		if topic == "" {
			topic = cfg.Ntfy.DefaultTopic
		}
		bridge := bridge.NewBridge(cfg.Ntfy.Server, topic, h)
		if !auth.IsEmpty() {
			bridge.WithAuth(auth)
		}
		http.Handle(route, bridge)
	}

	slog.Info("Server started", "address", cfg.HTTPAddress)
	fmt.Println(http.ListenAndServe(cfg.HTTPAddress, nil))
}
