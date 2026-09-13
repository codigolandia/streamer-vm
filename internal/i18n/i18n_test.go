package i18n

import (
	"encoding/json"
	"testing"
)

func TestLanguageResolution(t *testing.T) {
	// Test override
	if lang := resolveLanguage("pt"); lang != LangPT {
		t.Errorf("Expected pt for override 'pt', got %s", lang)
	}
	if lang := resolveLanguage("pt_BR"); lang != LangPT {
		t.Errorf("Expected pt for override 'pt_BR', got %s", lang)
	}
	if lang := resolveLanguage("en"); lang != LangEN {
		t.Errorf("Expected en for override 'en', got %s", lang)
	}
	if lang := resolveLanguage("fr"); lang != LangEN {
		t.Errorf("Expected en (fallback) for override 'fr', got %s", lang)
	}

	// Test STREAMER_LANG env
	t.Setenv("STREAMER_LANG", "pt_BR")
	if lang := resolveLanguage(""); lang != LangPT {
		t.Errorf("Expected pt for STREAMER_LANG=pt_BR, got %s", lang)
	}

	t.Setenv("STREAMER_LANG", "en_US")
	if lang := resolveLanguage(""); lang != LangEN {
		t.Errorf("Expected en for STREAMER_LANG=en_US, got %s", lang)
	}

	// Test LANG env
	t.Setenv("STREAMER_LANG", "")
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "pt_BR.UTF-8")
	if lang := resolveLanguage(""); lang != LangPT {
		t.Errorf("Expected pt for LANG=pt_BR.UTF-8, got %s", lang)
	}

	t.Setenv("LANG", "C.UTF-8")
	if lang := resolveLanguage(""); lang != LangEN {
		t.Errorf("Expected en for LANG=C.UTF-8, got %s", lang)
	}
}

func TestCatalogParity(t *testing.T) {
	enData, err := localesFS.ReadFile("locales/en.json")
	if err != nil {
		t.Fatalf("Failed to read en.json: %v", err)
	}
	ptData, err := localesFS.ReadFile("locales/pt.json")
	if err != nil {
		t.Fatalf("Failed to read pt.json: %v", err)
	}

	var enMap map[string]string
	if err := json.Unmarshal(enData, &enMap); err != nil {
		t.Fatalf("Failed to parse en.json: %v", err)
	}

	var ptMap map[string]string
	if err := json.Unmarshal(ptData, &ptMap); err != nil {
		t.Fatalf("Failed to parse pt.json: %v", err)
	}

	// Check that every key in enMap exists in ptMap
	for k := range enMap {
		if _, ok := ptMap[k]; !ok {
			t.Errorf("Key '%s' present in en.json but MISSING in pt.json", k)
		}
	}

	// Check that every key in ptMap exists in enMap
	for k := range ptMap {
		if _, ok := enMap[k]; !ok {
			t.Errorf("Key '%s' present in pt.json but MISSING in en.json", k)
		}
	}
}

func TestTranslation(t *testing.T) {
	Init("en")
	if val := T("cli.usage"); val != "Usage:" {
		t.Errorf("Expected 'Usage:', got '%s'", val)
	}

	Init("pt")
	if val := T("cli.usage"); val != "Uso:" {
		t.Errorf("Expected 'Uso:', got '%s'", val)
	}

	// Test formatting
	msg := T("msg.creating_vm", "test-vm", 2, 4, 20, 5900)
	if msg == "" || msg == "msg.creating_vm" {
		t.Errorf("Expected formatted translation, got '%s'", msg)
	}
}
