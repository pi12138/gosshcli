package i18n

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*
var locales embed.FS

var bundle *i18n.Bundle

var currentLang language.Tag

var supportedLangs = []language.Tag{
	language.AmericanEnglish,
	language.Chinese,
}

func init() {
	bundle = i18n.NewBundle(language.AmericanEnglish)

	if err := loadLocalizeFiles(); err != nil {
		fmt.Printf("Warning: failed to load localization files: %v\n", err)
	}

	currentLang = detectLanguage()
}

func loadLocalizeFiles() error {
	entries, err := locales.ReadDir("locales")
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join("locales", entry.Name())
			content, err := locales.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			if _, err := bundle.ParseMessageFileBytes(content, filePath); err != nil {
				return fmt.Errorf("failed to parse message file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

func detectLanguage() language.Tag {
	langEnv := os.Getenv("LANG")
	if langEnv != "" {
		langEnv = strings.Split(langEnv, ".")[0]
		langEnv = strings.Split(langEnv, "_")[0]
		tag, err := language.Parse(langEnv)
		if err == nil {
			if tag == language.Chinese || tag.String() == "zh" {
				return language.Chinese
			}
			if tag == language.AmericanEnglish || tag.String() == "en" {
				return language.AmericanEnglish
			}
		}
	}

	langEnv = os.Getenv("LANGUAGE")
	if langEnv != "" {
		langs := strings.Split(langEnv, ":")
		for _, lang := range langs {
			lang = strings.Split(lang, ".")[0]
			lang = strings.Split(lang, "_")[0]
			tag, err := language.Parse(lang)
			if err == nil {
				if tag == language.Chinese || lang == "zh" {
					return language.Chinese
				}
				if tag == language.AmericanEnglish || lang == "en" {
					return language.AmericanEnglish
				}
			}
		}
	}

	return language.AmericanEnglish
}

func GetCurrentLanguage() string {
	return currentLang.String()
}

func Localize(key string, templateData map[string]interface{}) string {
	localizer := i18n.NewLocalizer(bundle, GetCurrentLanguage())

	config := &i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	}

	str, err := localizer.Localize(config)
	if err != nil {
		return key
	}

	return str
}

func T(key string) string {
	return Localize(key, nil)
}

func TWith(key string, templateData map[string]interface{}) string {
	return Localize(key, templateData)
}

func Error(key string, templateData map[string]interface{}) error {
	return fmt.Errorf("%s", Localize(key, templateData))
}

func ErrorWith(key string, templateData map[string]interface{}, err error) error {
	return fmt.Errorf("%s: %w", Localize(key, templateData), err)
}
