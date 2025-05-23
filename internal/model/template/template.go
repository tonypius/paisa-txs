package template

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

//go:embed all:templates
var BuiltinTemplates embed.FS

type TemplateType string

const (
	Builtin TemplateType = "builtin"
	Custom  TemplateType = "custom"
)

type Rule struct {
	Condition string   `json:"condition"`
	Tags      []string `json:"tags"`
	Skip      bool     `json:"skip"`
}

type Template struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Content      string       `json:"content"`
	Rules        []Rule       `json:"rules"`
	TemplateType TemplateType `json:"template_type"`
}

func All() []Template {
	var templates []Template

	for _, t := range config.GetConfig().ImportTemplates {
		// Convert config.ImportRule to template.Rule
		var rules []Rule
		for _, r := range t.Rules {
			rules = append(rules, Rule{
				Condition: r.Condition,
				Tags:      r.Tags,
				Skip:      r.Skip,
			})
		}
		template := Template{
			ID:           buildID(t.Name, Custom),
			Name:         t.Name,
			Content:      t.Content,
			Rules:        rules,
			TemplateType: Custom,
		}
		templates = append(templates, template)
	}

	dirEntries, err := BuiltinTemplates.ReadDir("templates")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range dirEntries {
		name := f.Name()
		content, err := BuiltinTemplates.ReadFile(fmt.Sprintf("templates/%s", name))
		if err != nil {
			log.Fatal(err)
		}

		name = strings.TrimSuffix(name, filepath.Ext(name))
		template := Template{ID: buildID(name, Builtin), Name: name, Content: string(content), Rules: []Rule{}, TemplateType: Builtin}
		templates = append(templates, template)
	}

	return templates
}

func Upsert(name string, content string, rules []Rule) Template {
	template := Template{ID: buildID(name, Custom), Name: name, Content: content, Rules: rules, TemplateType: Custom}

	if config.GetConfig().Readonly {
		return template
	}

	Delete(name)
	cfg := config.GetConfig()
	
	// Convert template.Rule to config.ImportRule
	var configRules []config.ImportRule
	for _, r := range rules {
		configRules = append(configRules, config.ImportRule{
			Condition: r.Condition,
			Tags:      r.Tags,
			Skip:      r.Skip,
		})
	}
	
	cfg.ImportTemplates = append(cfg.ImportTemplates, config.ImportTemplate{
		Name:    name,
		Content: content,
		Rules:   configRules,
	})
	
	err := config.SaveConfigObject(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return template
}

func Delete(name string) {
	cfg := config.GetConfig()
	cfg.ImportTemplates = lo.Filter(cfg.ImportTemplates, func(t config.ImportTemplate, _ int) bool {
		return t.Name != name
	})

	err := config.SaveConfigObject(cfg)

	if err != nil {
		log.Fatal(err)
	}
}

func buildID(name string, templateType TemplateType) string {
	return fmt.Sprintf("%s:%s", templateType, name)
}
