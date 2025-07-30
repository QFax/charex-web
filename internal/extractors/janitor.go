package extractors

import (
	"charex/internal/core"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	// Regex to find the scenario tag.
	scenarioRegex = regexp.MustCompile(`(?s)<scenario>(.*?)<\/scenario>`)
	// Regex to find the example dialogs tag.
	exampleDialogsRegex = regexp.MustCompile(`(?s)<example_dialogs>(.*?)<\/example_dialogs>`)
	// Regex to find the name from a "Name:" key.
	nameKeyRegex = regexp.MustCompile(`(?i)Name:\s*([^\n\r]+)`)
)

// JanitorAIExtractor specializes in extracting character data from JanitorAI-style API request bodies.
type JanitorAIExtractor struct{}

// NewJanitorAIExtractor creates a new instance of the JanitorAIExtractor.
func NewJanitorAIExtractor() *JanitorAIExtractor {
	return &JanitorAIExtractor{}
}

// JAIMessage represents a single message in the JAI request.
type JAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Extract parses the JSON body of a JanitorAI request to create a character card.
func (e *JanitorAIExtractor) Extract(input []byte) (*core.TavernCardV2, []byte, []byte, error) {
	var messages []JAIMessage
	if err := json.Unmarshal(input, &messages); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to unmarshal janitorai request: %w", err)
	}

	// The raw data is the input byte slice itself.
	rawData := input

	// Concatenate all system prompts.
	var systemPromptBuilder strings.Builder
	for _, msg := range messages {
		if msg.Role == "system" {
			systemPromptBuilder.WriteString(msg.Content)
			systemPromptBuilder.WriteString("\n\n")
		}
	}
	systemPrompt := strings.TrimSpace(systemPromptBuilder.String())

	// Extract character details from the system prompt.
	charName, description, scenario, mesExample := extractDetailsFromSystemPrompt(systemPrompt)
	firstMes := extractFirstAssistantMessage(messages)
	userName := detectUserName(messages)

	// Anonymize the text fields.
	anonDesc := anonymizeText(description, charName, userName)
	anonScenario := anonymizeText(scenario, charName, userName)
	anonFirstMes := anonymizeText(firstMes, charName, userName)
	anonMesExample := anonymizeText(mesExample, charName, userName)

	cardData := core.TavernCardData{
		Name:                   charName,
		Description:            anonDesc,
		Scenario:               anonScenario,
		FirstMes:               anonFirstMes,
		MesExample:             anonMesExample,
		Tags:                   []string{"JanitorAI"},
		Creator:                "charex",
		CharacterVersion:       "1.0",
		Extensions:             make(map[string]interface{}),
		Personality:            "", // JAI format doesn't have these fields.
		CreatorNotes:           "",
		SystemPrompt:           "",
		PostHistoryInstructions: "",
		AlternateGreetings:     []string{},
	}

	card := &core.TavernCardV2{
		Spec:        "chara_card_v2",
		SpecVersion: "2.0",
		Data:        cardData,
		DisplayName: charName,
	}

	// The JanitorAI extractor does not handle images.
	return card, rawData, nil, nil
}

func extractDetailsFromSystemPrompt(prompt string) (name, description, scenario, mesExample string) {
	// Extract scenario.
	if matches := scenarioRegex.FindStringSubmatch(prompt); len(matches) > 1 {
		scenario = strings.TrimSpace(matches[1])
	}

	// Extract example dialogs.
	if matches := exampleDialogsRegex.FindStringSubmatch(prompt); len(matches) > 1 {
		mesExample = strings.TrimSpace(matches[1])
	}

	// The description is what's left after removing the other blocks.
	description = scenarioRegex.ReplaceAllString(prompt, "")
	description = exampleDialogsRegex.ReplaceAllString(description, "")
	description = strings.TrimSpace(description)

	// Heuristics to find the character name.
	// Heuristic 1: Look for "Name:".
	if matches := nameKeyRegex.FindStringSubmatch(description); len(matches) > 1 {
		name = strings.TrimSpace(matches[1])
	}

	// Fallback name if no other heuristic works.
	if name == "" {
		name = fmt.Sprintf("char_%x", md5.Sum([]byte(description)))
	}

	return
}

func extractFirstAssistantMessage(messages []JAIMessage) string {
	for _, msg := range messages {
		if msg.Role == "assistant" {
			return msg.Content
		}
	}
	return ""
}

func detectUserName(messages []JAIMessage) string {
	// A simple heuristic: find the first user message that contains a colon.
	for _, msg := range messages {
		if msg.Role == "user" {
			if parts := strings.SplitN(msg.Content, ":", 2); len(parts) > 1 {
				name := strings.TrimSpace(parts[0])
				// Avoid common placeholders.
				lowerName := strings.ToLower(name)
				if lowerName != "user" && lowerName != "you" {
					return name
				}
			}
		}
	}
	return ""
}

func anonymizeText(text, charName, userName string) string {
	anonymized := text
	if charName != "" {
		// Use a regex to replace whole words only.
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(charName) + `\b`)
		anonymized = re.ReplaceAllString(anonymized, "{{char}}")
	}
	if userName != "" {
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(userName) + `\b`)
		anonymized = re.ReplaceAllString(anonymized, "{{user}}")
	}
	return anonymized
}