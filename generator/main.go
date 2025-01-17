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
}

func main() {
	configs, err := loadConfigs("generator/configs")
	if err != nil {
		fmt.Printf("Error loading configs: %v\n", err)
		os.Exit(1)
	}

	tmpl, err := template.ParseFiles("generator/templates/token.go.tmpl")
	if err != nil {
		fmt.Printf("Error loading template: %v\n", err)
		os.Exit(1)
	}

	for lang, config := range configs {
		if err := generateFile(tmpl, lang, config); err != nil {
			fmt.Printf("Error generating %s: %v\n", lang, err)
			os.Exit(1)
		}
	}
}

func generateFile(tmpl *template.Template, lang string, config LanguageConfig) error {
	outDir := filepath.Join("./lang", lang)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	outFile := filepath.Join(outDir, lang+"_gen.go")
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, config)
}

func loadConfigs(configDir string) (map[string]LanguageConfig, error) {
	configs := make(map[string]LanguageConfig)
	
	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %v", err)
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
		configs[langCode] = config
	}
	
	if len(configs) == 0 {
		return nil, fmt.Errorf("no language configurations found in %s", configDir)
	}
	
	return configs, nil
}