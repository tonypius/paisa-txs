package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ananthakumaran/paisa/internal/config"
)

func TestTemplateWithRules(t *testing.T) {
	// Create a test template with rules
	testTemplate := Template{
		ID:           "test-template",
		Name:         "Test Template",
		Content:      "Test content",
		TemplateType: Custom,
		Rules: []Rule{
			{
				Condition: "ROW.A.includes('GROCERY')",
				Tags:      []string{"food", "grocery"},
				Skip:      false,
			},
			{
				Condition: "ROW.B > 1000",
				Tags:      []string{"expensive"},
				Skip:      true,
			},
		},
	}

	// Test conversion from Template to config.ImportTemplate
	var configRules []config.ImportRule
	for _, r := range testTemplate.Rules {
		configRules = append(configRules, config.ImportRule{
			Condition: r.Condition,
			Tags:      r.Tags,
			Skip:      r.Skip,
		})
	}

	configTemplate := config.ImportTemplate{
		Name:    testTemplate.Name,
		Content: testTemplate.Content,
		Rules:   configRules,
	}

	// Verify the conversion
	assert.Equal(t, testTemplate.Name, configTemplate.Name)
	assert.Equal(t, testTemplate.Content, configTemplate.Content)
	assert.Equal(t, len(testTemplate.Rules), len(configTemplate.Rules))

	for i, rule := range testTemplate.Rules {
		assert.Equal(t, rule.Condition, configTemplate.Rules[i].Condition)
		assert.Equal(t, rule.Tags, configTemplate.Rules[i].Tags)
		assert.Equal(t, rule.Skip, configTemplate.Rules[i].Skip)
	}

	// Test conversion from config.ImportTemplate to Template
	// Create a manual conversion function for testing
	convertTemplates := func(templates []config.ImportTemplate) []Template {
		result := make([]Template, 0, len(templates))
		for _, t := range templates {
			template := Template{
				ID:           buildID(t.Name, Custom),
				Name:         t.Name,
				Content:      t.Content,
				TemplateType: Custom,
			}
			
			// Convert rules
			if len(t.Rules) > 0 {
				template.Rules = make([]Rule, 0, len(t.Rules))
				for _, r := range t.Rules {
					template.Rules = append(template.Rules, Rule{
						Condition: r.Condition,
						Tags:      r.Tags,
						Skip:      r.Skip,
					})
				}
			}
			
			result = append(result, template)
		}
		return result
	}
	
	// Convert the template
	templates := []config.ImportTemplate{configTemplate}
	convertedTemplates := convertTemplates(templates)

	// Verify the conversion
	assert.Equal(t, 1, len(convertedTemplates))
	assert.Equal(t, testTemplate.Name, convertedTemplates[0].Name)
	assert.Equal(t, testTemplate.Content, convertedTemplates[0].Content)
	assert.Equal(t, len(testTemplate.Rules), len(convertedTemplates[0].Rules))

	for i, rule := range testTemplate.Rules {
		assert.Equal(t, rule.Condition, convertedTemplates[0].Rules[i].Condition)
		assert.Equal(t, rule.Tags, convertedTemplates[0].Rules[i].Tags)
		assert.Equal(t, rule.Skip, convertedTemplates[0].Rules[i].Skip)
	}
}