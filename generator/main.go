package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"gopkg.in/yaml.v2"
)

type LanguageConfig struct {
	Code string
	Name string
	IsIndic bool
}

var IndicLangs = []string{
	"hin", "ben", "fas", "guj", "mar", "pan", "sin", "urd", "tam", "tel",
}

func main() {
	configs, err := loadConfigs("generator/configs")
	if err != nil {
		fmt.Printf("Error loading configs: %v\n", err)
		os.Exit(1)
	}

	// Load both templates
	tmpl, err := template.ParseFiles(
		"generator/templates/token.go.tmpl",
		"generator/templates/init.go.tmpl",
	)
	if err != nil {
		fmt.Printf("Error loading templates: %v\n", err)
		os.Exit(1)
	}

	for lang, config := range configs {
		if err := generateFiles(tmpl, lang, config); err != nil {
			fmt.Printf("Error generating %s: %v\n", lang, err)
			os.Exit(1)
		}
	}
}

func generateFiles(tmpl *template.Template, lang string, config LanguageConfig) error {
	outDir := filepath.Join("./lang", lang)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Generate token_gen.go
	if err := generateFile(tmpl, "token.go.tmpl", filepath.Join(outDir, lang+"_gen.go"), config); err != nil {
		return err
	}

	// Generate init_gen.go for Indic languages
	if isIndicLanguage(lang) {
		if err := generateFile(tmpl, "init.go.tmpl", filepath.Join(outDir, "init_gen.go"), config); err != nil {
			return err
		}
	}

	return nil
}

func generateFile(tmpl *template.Template, templateName, outFile string, config LanguageConfig) error {
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.ExecuteTemplate(f, templateName, config)
}

func loadConfigs(configDir string) (map[string]LanguageConfig, error) {
	configs := make(map[string]LanguageConfig)
	
	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}
	
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}
		
		langCode := strings.TrimSuffix(strings.TrimSuffix(file.Name(), ".yaml"), ".yml")
		
		content, err := os.ReadFile(filepath.Join(configDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %v", file.Name(), err)
		}
		
		var config LanguageConfig
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %v", file.Name(), err)
		}
		
		config.Code = langCode
		config.IsIndic = isIndicLanguage(langCode)
		
		configs[langCode] = config
	}
	
	if len(configs) == 0 {
		return nil, fmt.Errorf("no language configurations found in %s", configDir)
	}
	
	return configs, nil
}

func isIndicLanguage(lang string) bool {
	for _, indicLang := range IndicLangs {
		if lang == indicLang {
			return true
		}
	}
	return false
}
