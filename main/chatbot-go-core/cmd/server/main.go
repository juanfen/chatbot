package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"xiaozhi-go-core/internal/config"
	"xiaozhi-go-core/internal/gateway"
	"xiaozhi-go-core/internal/observability"
	"xiaozhi-go-core/internal/service"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	metricsAddr := flag.String("metrics_addr", ":9090", "metrics http listen address")
	flag.Parse()

	logger := observability.NewLogger()

	cfg, err := config.LoadFromFile(*configPath)
	if err != nil {
		logger.Error("load config failed", "error", err)
		os.Exit(1)
	}
	cfgStore := config.NewStore(cfg)

	metrics := observability.NewMetrics(logger)
	stopMetrics := make(chan struct{})
	go metrics.LogPeriodic(60*time.Second, stopMetrics)

	chatService := service.NewChatService(cfgStore, logger)
	voiceService := service.NewVoicePipelineService(cfgStore, chatService, logger)
	wsServer := gateway.NewWSServer(cfgStore, voiceService, logger)
	httpGateway := gateway.NewHTTPServer(cfgStore, logger)
	startMetricsHTTP(*metricsAddr, metrics, chatService, logger)
	wsServer.Start()
	httpGateway.Start()
	done := make(chan struct{})

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		logger.Info("received signal, shutting down", "signal", sig)
		close(stopMetrics)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		wsServer.Stop(shutdownCtx)
		httpGateway.Stop(shutdownCtx)
		close(done)
	}()

	printStartupSummary(logger, cfgStore.Get(), *configPath, *metricsAddr)

	app := cfgStore.Get()
	logger.Info("starting xiaozhi-go-core",
		"config_file", *configPath,
		"metrics_addr", *metricsAddr,
		"llm_model", app.LLM.ModelName,
		"max_concurrency", app.Server.MaxConcurrency,
	)
	<-done
}

func printStartupSummary(logger *slog.Logger, cfg *config.AppConfig, configPath, metricsAddr string) {
	cfgJSON, err := config.ToLogJSON(cfg)
	if err != nil {
		logger.Warn("startup config json failed", "error", err)
	} else {
		fmt.Printf("\n== XIAOZHI GO CORE CONFIG ==\nconfig_file: %s\n%s\n", configPath, string(cfgJSON))
	}

	httpBase := metricsHTTPBase(metricsAddr)
	httpRoutes := []string{
		"GET " + httpBase + "/metrics",
		"GET " + httpBase + "/status",
	}
	gatewayRoutes := []string{
		"WS " + cfg.Gateway.WSListenAddr + "/xiaozhi/v1/",
		"GET/POST " + cfg.Gateway.HTTPListenAddr + "/xiaozhi/ota/",
		"GET " + cfg.Gateway.HTTPListenAddr + "/xiaozhi/ota/download/{filename}",
		"GET/POST " + cfg.Gateway.HTTPListenAddr + "/xiaozhi/v1/runtime/config (Bearer XIAOZHI_GO_RUNTIME_CONFIG_TOKEN)",
	}

	fmt.Printf(
		"\n== XIAOZHI GO CORE ENDPOINTS ==\n"+
			"metrics_listen: %s\n"+
			"http_routes:\n%s\n"+
			"gateway_routes:\n%s\n\n",
		metricsAddrOrDisabled(metricsAddr),
		formatList(httpRoutes),
		formatList(gatewayRoutes),
	)
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "  - (none)"
	}
	var b strings.Builder
	for _, item := range items {
		b.WriteString("  - ")
		b.WriteString(item)
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func metricsAddrOrDisabled(addr string) string {
	if addr == "" {
		return "(disabled)"
	}
	return addr
}

func metricsHTTPBase(addr string) string {
	if addr == "" {
		return ""
	}
	host := addr
	if len(addr) > 0 && addr[0] == ':' {
		host = "127.0.0.1" + addr
	}
	return "http://" + host
}

func startMetricsHTTP(
	addr string,
	metrics *observability.Metrics,
	chatService *service.ChatService,
	logger *slog.Logger,
) {
	if addr == "" {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(metrics.RenderPrometheus()))
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := observabilityStatusJSON(chatService, metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(data)
	})

	go func() {
		logger.Info("metrics http listening", "addr", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			logger.Error("metrics http server error", "error", err)
		}
	}()
}

type statusResponse struct {
	Inflight int64                  `json:"inflight"`
	Metrics  observability.Snapshot `json:"metrics"`
}

func observabilityStatusJSON(chatService *service.ChatService, metrics *observability.Metrics) ([]byte, error) {
	resp := statusResponse{
		Inflight: chatService.Inflight(),
		Metrics:  metrics.Snapshot(),
	}
	return json.Marshal(resp)
}
