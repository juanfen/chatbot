package gateway

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"xiaozhi-go-core/internal/config"
)

type HTTPServer struct {
	store  *config.Store
	logger *slog.Logger
	server *http.Server
	ota    *OTAHandler
}

func NewHTTPServer(store *config.Store, logger *slog.Logger) *HTTPServer {
	return &HTTPServer{
		store:  store,
		logger: logger,
		ota:    NewOTAHandler(store, logger),
	}
}

func (s *HTTPServer) Start() {
	gw := s.store.Get().Gateway
	if !gw.Enabled || gw.HTTPListenAddr == "" {
		return
	}
	mux := http.NewServeMux()
	s.registerRuntimeConfigRoutes(mux)
	s.ota.Register(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			_, _ = w.Write([]byte("Server is running"))
			return
		}
		http.NotFound(w, r)
	})
	s.server = &http.Server{
		Addr:              gw.HTTPListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		s.logger.Info("gateway_http_listening", "addr", gw.HTTPListenAddr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("gateway_http_error", "error", err)
		}
	}()
}

func (s *HTTPServer) Stop(ctx context.Context) {
	if s.server != nil {
		_ = s.server.Shutdown(ctx)
	}
}

// registerRuntimeConfigRoutes 注册运行时配置查询与补丁接口（不写磁盘）。
func (s *HTTPServer) registerRuntimeConfigRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/xiaozhi/v1/runtime/config", s.handleRuntimeConfig)
}

func (s *HTTPServer) handleRuntimeConfig(w http.ResponseWriter, r *http.Request) {
	addCORS(w)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Device-Id, Client-Id")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodGet:
		view := config.RuntimeSectionsFrom(s.store.Get())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(view); err != nil {
			s.logger.Warn("runtime_config_get_encode_failed", "error", err)
		}
	case http.MethodPost:
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<22))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.store.ApplyRuntimePatchJSON(body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.logger.Info("runtime_config_patch_applied", "patch", string(body))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
