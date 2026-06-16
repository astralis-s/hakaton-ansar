// Package config загружает настройки бота из окружения (и опционально из .env).
package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config — все настройки сервиса. Секретов в коде нет, только из окружения.
type Config struct {
	BotToken        string        // токен @BotFather (обязателен)
	HTTPAddr        string        // адрес веб-инбокса менеджера
	ManagerPassword string        // пароль входа в инбокс ("" — без пароля)
	DataFile        string        // путь к JSON-хранилищу
	PollTimeout     time.Duration // таймаут long polling
}

// Load читает конфиг из переменных окружения, подставляя разумные значения по
// умолчанию. Единственное обязательное поле — BOT_TOKEN.
func Load() (Config, error) {
	cfg := Config{
		BotToken:        strings.TrimSpace(os.Getenv("BOT_TOKEN")),
		HTTPAddr:        envOr("HTTP_ADDR", ":8090"),
		ManagerPassword: os.Getenv("MANAGER_PASSWORD"),
		DataFile:        envOr("DATA_FILE", "data/bot-data.json"),
		PollTimeout:     time.Duration(envInt("POLL_TIMEOUT_SEC", 30)) * time.Second,
	}
	if cfg.BotToken == "" {
		return Config{}, errors.New("BOT_TOKEN не задан (получите его у @BotFather и положите в .env)")
	}
	if cfg.PollTimeout <= 0 {
		cfg.PollTimeout = 30 * time.Second
	}
	return cfg, nil
}

// LoadDotenv делает best-effort загрузку KEY=VALUE из .env-файла в окружение.
// Уже установленные переменные не перезаписываются. Отсутствие файла — не ошибка.
func LoadDotenv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
