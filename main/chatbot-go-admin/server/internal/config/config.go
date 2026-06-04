package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config is the application runtime configuration.
type Config struct {
	HTTPAddr  string
	LogLevel  string
	JWTSecret string
	SchemaDir string
	DB        DBConfig

	// XiaoZhiCoreBaseURL 指向 xiaozhi-go-core 的 Gateway HTTP 监听地址
	// 例如：http://127.0.0.1:8003 ；admin 通过反向代理把
	// 「算法运行时配置」前端请求转给该服务（避免浏览器 CORS / 集中加鉴权）。
	XiaoZhiCoreBaseURL string

	// AudioTargetPath 为语音样本落盘目录（相对路径相对于进程工作目录；建议绝对路径）。
	// 上传后统一转码为 24kHz 单声道 PCM WAV，文件名为随机 UUID。
	AudioTargetPath string
	FFmpegBin       string
	FFprobeBin      string
	// AudioDenoise 上传转码时是否启用 ffmpeg 语音降噪（高通/低通 + afftdn + 压缩 + 响度归一）。
	AudioDenoise bool

	// CORSOrigins 额外允许的浏览器 Origin（逗号分隔），用于本机前端直连远程 API。
	// 示例：http://localhost:9527,http://127.0.0.1:9527
	CORSOrigins []string

	// SV 声纹服务：分配音色时向注册接口申报（multipart: speaker_id + file）
	SVVoiceprintRegisterURL  string
	SVVoiceprintIdentifyURL  string // 兼容旧配置，未设 RegisterURL 时自动把 identify 换成 register
	SVVoiceprintToken        string
	SVVoiceprintTimeoutSec   int

	// MemoryProfileBaseURL 长期记忆画像服务，例如 http://192.168.99.103:8176
	MemoryProfileBaseURL     string
	MemoryProfileTimeoutSec  int
}

// ResolveSVVoiceprintRegisterURL returns the voiceprint register endpoint.
func (c Config) ResolveSVVoiceprintRegisterURL() string {
	if u := strings.TrimSpace(c.SVVoiceprintRegisterURL); u != "" {
		return u
	}
	u := strings.TrimSpace(c.SVVoiceprintIdentifyURL)
	if u == "" {
		return ""
	}
	if strings.HasSuffix(u, "/identify") {
		return strings.TrimSuffix(u, "/identify") + "/register"
	}
	if strings.Contains(u, "/identify") {
		return strings.Replace(u, "/identify", "/register", 1)
	}
	return u
}

// DBConfig holds database connection settings.
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Load loads configuration from environment variables with safe defaults for local development.
//
// 加载顺序（先到先得，**已存在的真实环境变量永远优先**）：
//  1. 进程环境变量（命令行 / 系统 export）
//  2. ADMIN_DOTENV_PATH 显式指定的文件
//  3. 当前工作目录下 ./.env
//  4. 可执行文件同目录下的 .env
//
// godotenv.Load 的默认语义就是「不覆盖已存在的 env」，因此外部传入的环境变量
// 始终拥有最高优先级，与 12-factor 风格一致。
func Load() Config {
	loadDotEnvFiles()

	return Config{
		HTTPAddr:           getEnv("ADMIN_HTTP_ADDR", ":8088"),
		LogLevel:           getEnv("ADMIN_LOG_LEVEL", "info"),
		JWTSecret:          getEnv("ADMIN_JWT_SECRET", "dev-secret-change-me"),
		SchemaDir:          getEnv("ADMIN_SCHEMA_DIR", "resources/schema"),
		XiaoZhiCoreBaseURL: getEnv("XIAOZHI_CORE_BASE_URL", "http://127.0.0.1:8003"),
		AudioTargetPath:    getEnv("ADMIN_AUDIO_TARGET_PATH", "data/audio"),
		FFmpegBin:          getEnv("ADMIN_FFMPEG_BIN", "ffmpeg"),
		FFprobeBin:         getEnv("ADMIN_FFPROBE_BIN", "ffprobe"),
		AudioDenoise:       getEnvBool("ADMIN_AUDIO_DENOISE", true),
		CORSOrigins:        splitCSV(getEnv("ADMIN_CORS_ORIGINS", "")),
		SVVoiceprintRegisterURL:  getEnv("SV_VOICEPRINT_REGISTER_URL", ""),
		SVVoiceprintIdentifyURL:  getEnv("SV_VOICEPRINT_IDENTIFY_URL", ""),
		SVVoiceprintToken:        getEnv("SV_VOICEPRINT_TOKEN", ""),
		SVVoiceprintTimeoutSec:   getEnvInt("SV_VOICEPRINT_TIMEOUT_SEC", 120),
		MemoryProfileBaseURL:     getEnv("MEMORY_PROFILE_BASE_URL", ""),
		MemoryProfileTimeoutSec:  getEnvInt("MEMORY_PROFILE_TIMEOUT_SEC", 30),
		DB: DBConfig{
			Host:     getEnv("PG_HOST", "127.0.0.1"),
			Port:     getEnvInt("PG_PORT", 5432),
			User:     getEnv("PG_USER", "postgres"),
			Password: getEnv("PG_PASSWORD", "postgres"),
			DBName:   getEnv("PG_DBNAME", "postgres"),
			SSLMode:  getEnv("PG_SSLMODE", "disable"),
		},
	}
}

// ResolveAudioDir returns an absolute path for storing converted voice WAV files.
func (c Config) ResolveAudioDir() (string, error) {
	p := c.AudioTargetPath
	if p == "" {
		p = "data/audio"
	}
	return filepath.Abs(p)
}

// loadDotEnvFiles 尝试从一组候选路径加载 .env 文件。
// 找不到/解析失败都不会让进程崩溃，仅静默跳过。
func loadDotEnvFiles() {
	var candidates []string

	if p := os.Getenv("ADMIN_DOTENV_PATH"); p != "" {
		candidates = append(candidates, p)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, ".env"))
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, p := range candidates {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			abs = p
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		if info, err := os.Stat(abs); err != nil || info.IsDir() {
			continue
		}
		// godotenv.Load 默认不会覆盖已存在的环境变量，
		// 所以多个 .env 之间也是「先到先得」。
		_ = godotenv.Load(abs)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnvBool(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
