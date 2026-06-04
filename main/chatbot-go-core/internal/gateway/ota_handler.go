package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"xiaozhi-go-core/internal/config"
)

type OTAHandler struct {
	store     *config.Store
	logger    *slog.Logger
	mu        sync.Mutex
	cacheAt   time.Time
	firmwares map[string][]firmwareFile
}

type firmwareFile struct {
	Model    string
	Version  string
	Filename string
}

func NewOTAHandler(store *config.Store, logger *slog.Logger) *OTAHandler {
	return &OTAHandler{
		store:     store,
		logger:    logger,
		firmwares: map[string][]firmwareFile{},
	}
}

func (h *OTAHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/xiaozhi/ota/", h.handleOTA)
	mux.HandleFunc("/xiaozhi/ota/download/", h.handleDownload)
}

func (h *OTAHandler) handleOTA(w http.ResponseWriter, r *http.Request) {
	addCORS(w)
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Device-Id, Client-Id")
		w.WriteHeader(http.StatusOK)
		return
	}
	cfg := h.store.Get()
	if !cfg.OTA.Enabled {
		h.writeJSON(w, map[string]any{"success": false, "message": "ota disabled"})
		return
	}
	if r.Method == http.MethodGet {
		h.writeText(w, fmt.Sprintf("OTA接口运行正常，向设备发送的websocket地址是：%s", h.websocketURL(cfg, r)))
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deviceID := r.Header.Get("device-id")
	clientID := r.Header.Get("client-id")
	if deviceID == "" || clientID == "" {
		h.writeJSON(w, map[string]any{"success": false, "message": "request error."})
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)

	model := firstNonEmpty(r.Header.Get("device-model"), r.Header.Get("device_model"), r.Header.Get("model"), stringFromPath(body, "board", "type"), stringValue(body["model"]), "default")
	version := firstNonEmpty(r.Header.Get("device-version"), r.Header.Get("device_version"), r.Header.Get("firmware-version"), r.Header.Get("app-version"), r.Header.Get("application-version"), stringFromPath(body, "application", "version"), "0.0.0")

	resp := map[string]any{
		"server_time": map[string]any{
			"timestamp":       time.Now().UnixMilli(),
			"timezone_offset": 8 * 60,
		},
		"firmware": map[string]any{
			"version": version,
			"url":     "",
		},
	}

	if cfg.Gateway.MQTTGateway != "" {
		resp["mqtt"] = h.mqttConfig(cfg, deviceID, clientID, model)
	} else {
		token := ""
		if cfg.Gateway.Auth.Enabled && !h.isAllowedDevice(cfg, deviceID) {
			auth := NewAuthManager(cfg.Gateway.AuthKey, cfg.Gateway.Auth.ExpireSeconds)
			token = auth.GenerateToken(clientID, deviceID)
		}
		resp["websocket"] = map[string]any{
			"url":   h.websocketURL(cfg, r),
			"token": token,
		}
	}

	if fw := h.findUpgrade(cfg, model, version); fw != nil {
		resp["firmware"] = map[string]any{
			"version": fw.Version,
			"url":     h.downloadURL(cfg, r, fw.Filename),
		}
	}
	h.writeJSON(w, resp)
}

func (h *OTAHandler) handleDownload(w http.ResponseWriter, r *http.Request) {
	addCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := h.store.Get()
	if !cfg.OTA.Enabled {
		http.Error(w, "ota disabled", http.StatusNotFound)
		return
	}
	name := filepath.Base(strings.TrimPrefix(r.URL.Path, "/xiaozhi/ota/download/"))
	if !regexp.MustCompile(`^[A-Za-z0-9._-]+\.bin$`).MatchString(name) {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}
	path := filepath.Join(cfg.OTA.BinDir, name)
	realPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	binDir, _ := filepath.Abs(cfg.OTA.BinDir)
	if !strings.HasPrefix(realPath, binDir+string(os.PathSeparator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	http.ServeFile(w, r, realPath)
}

func (h *OTAHandler) websocketURL(cfg *config.AppConfig, r *http.Request) string {
	if cfg.Gateway.WebSocketURL != "" && !strings.Contains(cfg.Gateway.WebSocketURL, "你的") {
		return cfg.Gateway.WebSocketURL
	}
	host := r.Host
	if cfg.Gateway.WSListenAddr != "" {
		host = hostWithPort(r.Host, cfg.Gateway.WSListenAddr)
	}
	return "ws://" + host + "/xiaozhi/v1/"
}

func (h *OTAHandler) downloadURL(cfg *config.AppConfig, r *http.Request, filename string) string {
	host := hostWithPort(r.Host, cfg.Gateway.HTTPListenAddr)
	return "http://" + host + "/xiaozhi/ota/download/" + filename
}

func (h *OTAHandler) mqttConfig(cfg *config.AppConfig, deviceID string, clientID string, model string) map[string]any {
	groupID := strings.NewReplacer(":", "_", " ", "_").Replace("GID_" + model)
	mac := strings.ReplaceAll(deviceID, ":", "_")
	mqttClientID := groupID + "@@@" + mac + "@@@" + mac
	userData, _ := json.Marshal(map[string]string{"ip": "unknown"})
	username := base64.StdEncoding.EncodeToString(userData)
	password := ""
	if cfg.Gateway.AuthKey != "" {
		mac := hmac.New(sha256.New, []byte(cfg.Gateway.AuthKey))
		_, _ = mac.Write([]byte(mqttClientID + "|" + username))
		password = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}
	return map[string]any{
		"endpoint":        cfg.Gateway.MQTTGateway,
		"client_id":       mqttClientID,
		"username":        username,
		"password":        password,
		"publish_topic":   "device-server",
		"subscribe_topic": "devices/p2p/" + mac,
	}
}

func (h *OTAHandler) findUpgrade(cfg *config.AppConfig, model string, current string) *firmwareFile {
	h.refreshFirmwareCache(cfg)
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, fw := range h.firmwares[model] {
		if versionGreater(fw.Version, current) {
			cp := fw
			return &cp
		}
	}
	return nil
}

func (h *OTAHandler) refreshFirmwareCache(cfg *config.AppConfig) {
	ttl := time.Duration(cfg.OTA.FirmwareCacheTTL) * time.Second
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	h.mu.Lock()
	if time.Since(h.cacheAt) < ttl && len(h.firmwares) > 0 {
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	matches, _ := filepath.Glob(filepath.Join(cfg.OTA.BinDir, "*.bin"))
	next := map[string][]firmwareFile{}
	re := regexp.MustCompile(`^(.+?)_([0-9][A-Za-z0-9._-]*)\.bin$`)
	for _, path := range matches {
		name := filepath.Base(path)
		parts := re.FindStringSubmatch(name)
		if len(parts) != 3 {
			continue
		}
		next[parts[1]] = append(next[parts[1]], firmwareFile{Model: parts[1], Version: parts[2], Filename: name})
	}
	for model := range next {
		sort.Slice(next[model], func(i, j int) bool {
			return versionGreater(next[model][i].Version, next[model][j].Version)
		})
	}
	h.mu.Lock()
	h.firmwares = next
	h.cacheAt = time.Now()
	h.mu.Unlock()
}

func (h *OTAHandler) isAllowedDevice(cfg *config.AppConfig, deviceID string) bool {
	for _, allowed := range cfg.Gateway.Auth.AllowedDevices {
		if allowed == deviceID {
			return true
		}
	}
	return false
}

func (h *OTAHandler) writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(payload)
	_, _ = w.Write(data)
}

func (h *OTAHandler) writeText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(text))
}

func addCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Device-Id, Client-Id")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func stringFromPath(m map[string]any, keys ...string) string {
	var cur any = m
	for _, key := range keys {
		next, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = next[key]
	}
	return stringValue(cur)
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func hostWithPort(host string, listenAddr string) string {
	port := strings.TrimPrefix(listenAddr, ":")
	if strings.Contains(listenAddr, ":") && !strings.HasPrefix(listenAddr, ":") {
		_, p, err := strings.Cut(listenAddr, ":")
		if err {
			port = p
		}
	}
	if h, _, ok := strings.Cut(host, ":"); ok {
		host = h
	}
	return host + ":" + port
}

func versionGreater(a string, b string) bool {
	av := parseVersion(a)
	bv := parseVersion(b)
	n := len(av)
	if len(bv) > n {
		n = len(bv)
	}
	for i := 0; i < n; i++ {
		var ai, bi int
		if i < len(av) {
			ai = av[i]
		}
		if i < len(bv) {
			bi = bv[i]
		}
		if ai != bi {
			return ai > bi
		}
	}
	return false
}

func parseVersion(v string) []int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(v, -1)
	if len(matches) == 0 {
		return []int{0}
	}
	out := make([]int, 0, len(matches))
	for _, match := range matches {
		n, _ := strconv.Atoi(match)
		out = append(out, n)
	}
	return out
}
