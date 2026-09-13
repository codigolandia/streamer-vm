package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localesFS embed.FS

// Language represents a supported language code.
type Language string

const (
	LangEN Language = "en"
	LangPT Language = "pt"
)

var (
	mu              sync.RWMutex
	currentLang     = LangEN
	activeCatalog   = map[string]string{}
	fallbackCatalog = map[string]string{}
)

func init() {
	// Initialize with environment detection by default
	Init("")
}

// Init initializes the i18n subsystem.
// Language priority: override > STREAMER_LANG > LC_ALL > LC_MESSAGES > LANG > LangEN (default).
func Init(override string) {
	mu.Lock()
	defer mu.Unlock()

	// Always load English as fallback catalog
	fallbackCatalog = loadCatalogFile(LangEN)

	lang := resolveLanguage(override)
	currentLang = lang

	if lang == LangEN {
		activeCatalog = fallbackCatalog
	} else {
		activeCatalog = loadCatalogFile(lang)
	}
}

// resolveLanguage determines the language code based on parameters and environment.
func resolveLanguage(override string) Language {
	if override != "" {
		if strings.HasPrefix(strings.ToLower(override), "pt") {
			return LangPT
		}
		return LangEN
	}

	if env := os.Getenv("STREAMER_LANG"); env != "" {
		if strings.HasPrefix(strings.ToLower(env), "pt") {
			return LangPT
		}
		return LangEN
	}

	for _, envKey := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if val := os.Getenv(envKey); val != "" {
			lower := strings.ToLower(val)
			if strings.HasPrefix(lower, "pt") {
				return LangPT
			}
			break
		}
	}

	return LangEN
}

func loadCatalogFile(lang Language) map[string]string {
	path := fmt.Sprintf("locales/%s.json", lang)
	data, err := localesFS.ReadFile(path)
	if err != nil {
		return map[string]string{}
	}

	cat := make(map[string]string)
	_ = json.Unmarshal(data, &cat)
	return cat
}

// T translates a message key and formats any parameters using fmt.Sprintf.
func T(key string, args ...any) string {
	mu.RLock()
	defer mu.RUnlock()

	val, ok := activeCatalog[key]
	if !ok || val == "" {
		val, ok = fallbackCatalog[key]
		if !ok || val == "" {
			val = key
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

// Current returns the currently active language.
func Current() Language {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}
