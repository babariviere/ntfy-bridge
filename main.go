package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"forge.babariviere.com/babariviere/ntfy-bridge/bridge"
	"forge.babariviere.com/babariviere/ntfy-bridge/config"
)

func main() {
	cfg, err := config.ReadConfig("config.scfg")
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
		}

		slog.Debug("Registering bridge", "route", route, "handler", handler.Type)
		bridge := bridge.NewBridge(cfg.Ntfy.Server, handler.Topic, h)
		if !auth.IsEmpty() {
			bridge.WithAuth(auth)
		}
		http.Handle(route, bridge)
	}

	slog.Info("Server started", "address", cfg.HTTPAddress)
	fmt.Println(http.ListenAndServe(cfg.HTTPAddress, nil))
}
