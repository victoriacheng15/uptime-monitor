package web

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Generator struct {
	TemplatesDir string
	ContentDir   string
	OutputDir    string
	APIBaseURL   string
}

func NewGenerator(templatesDir, contentDir, outputDir string) *Generator {
	return &Generator{
		TemplatesDir: templatesDir,
		ContentDir:   contentDir,
		OutputDir:    outputDir,
	}
}

func (g *Generator) Generate() error {
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return err
	}

	landingConfig, err := g.loadLandingConfig()
	if err != nil {
		return err
	}

	evolutionConfig, err := g.loadEvolutionConfig()
	if err != nil {
		return err
	}

	g.reverseEvolution(evolutionConfig)

	currentYear := time.Now().Year()

	data := TemplateData{
		Landing:    landingConfig,
		Evolution:  evolutionConfig,
		Year:       currentYear,
		APIBaseURL: g.APIBaseURL,
	}

	if err := g.renderWithBase("index.html", "index.html", data); err != nil {
		return err
	}

	if err := g.renderWithBase("monitor.html", "monitor.html", data); err != nil {
		return err
	}

	if err := g.renderWithBase("evolution.html", "evolution.html", data); err != nil {
		return err
	}

	if err := g.renderRawTemplate("llms.txt", "llms.txt", data); err != nil {
		return err
	}

	return nil
}

func (g *Generator) reverseEvolution(cfg *EvolutionConfig) {
	for i, j := 0, len(cfg.Chapters)-1; i < j; i, j = i+1, j-1 {
		cfg.Chapters[i], cfg.Chapters[j] = cfg.Chapters[j], cfg.Chapters[i]
	}
	for i := range cfg.Chapters {
		timeline := cfg.Chapters[i].Timeline
		for a, b := 0, len(timeline)-1; a < b; a, b = a+1, b-1 {
			timeline[a], timeline[b] = timeline[b], timeline[a]
		}
		cfg.Chapters[i].Timeline = timeline
	}
}

func (g *Generator) loadLandingConfig() (*SiteConfig, error) {
	data, err := os.ReadFile(filepath.Join(g.ContentDir, "landing.yaml"))
	if err != nil {
		return nil, err
	}
	var config SiteConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (g *Generator) loadEvolutionConfig() (*EvolutionConfig, error) {
	data, err := os.ReadFile(filepath.Join(g.ContentDir, "evolution.yaml"))
	if err != nil {
		return nil, err
	}
	var config EvolutionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (g *Generator) getFuncs() template.FuncMap {
	return template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"list": func(values ...interface{}) []interface{} {
			return values
		},
	}
}

func (g *Generator) renderWithBase(tmplName, outName string, data interface{}) error {
	basePath := filepath.Join(g.TemplatesDir, "base.html")
	tmplPath := filepath.Join(g.TemplatesDir, tmplName)

	tmpl, err := template.New("base.html").Funcs(g.getFuncs()).ParseFiles(basePath, tmplPath)
	if err != nil {
		return err
	}

	outFile, err := os.Create(filepath.Join(g.OutputDir, outName))
	if err != nil {
		return err
	}
	defer outFile.Close()

	return tmpl.ExecuteTemplate(outFile, "base.html", data)
}

func (g *Generator) renderRawTemplate(tmplName, outName string, data interface{}) error {
	tmplPath := filepath.Join(g.TemplatesDir, tmplName)
	tmpl, err := template.New(tmplName).Funcs(g.getFuncs()).ParseFiles(tmplPath)
	if err != nil {
		return err
	}

	outFile, err := os.Create(filepath.Join(g.OutputDir, outName))
	if err != nil {
		return err
	}
	defer outFile.Close()

	return tmpl.Execute(outFile, data)
}
