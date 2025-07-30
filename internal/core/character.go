package core

// TavernCardV2 represents the full V2 character card structure.
type TavernCardV2 struct {
	Spec        string         `json:"spec"`
	SpecVersion string         `json:"spec_version"`
	Data        TavernCardData `json:"data"`
	DisplayName string         `json:"-"` // Internal field for filename generation, not part of the JSON spec.
}

// TavernCardData contains the core character information.
type TavernCardData struct {
	Name                   string                 `json:"name"`
	Description            string                 `json:"description"`
	Personality            string                 `json:"personality"`
	Scenario               string                 `json:"scenario"`
	FirstMes               string                 `json:"first_mes"`
	MesExample             string                 `json:"mes_example"`
	CreatorNotes           string                 `json:"creator_notes"`
	SystemPrompt           string                 `json:"system_prompt"`
	PostHistoryInstructions string                 `json:"post_history_instructions"`
	AlternateGreetings     []string               `json:"alternate_greetings"`
	CharacterBook          *CharacterBook         `json:"character_book,omitempty"`
	Tags                   []string               `json:"tags"`
	Creator                string                 `json:"creator"`
	CharacterVersion       string                 `json:"character_version"`
	Extensions             map[string]interface{} `json:"extensions"`
}

// CharacterBook represents a character-specific lorebook.
type CharacterBook struct {
	Name              string                 `json:"name,omitempty"`
	Description       string                 `json:"description,omitempty"`
	ScanDepth         int                    `json:"scan_depth,omitempty"`
	TokenBudget       int                    `json:"token_budget,omitempty"`
	RecursiveScanning bool                   `json:"recursive_scanning,omitempty"`
	Extensions        map[string]interface{} `json:"extensions"`
	Entries           []BookEntry            `json:"entries"`
}

// BookEntry is an entry in a CharacterBook.
type BookEntry struct {
	Keys           []string               `json:"keys"`
	Content        string                 `json:"content"`
	Extensions     map[string]interface{} `json:"extensions"`
	Enabled        bool                   `json:"enabled"`
	InsertionOrder int                    `json:"insertion_order"`
	CaseSensitive  bool                   `json:"case_sensitive,omitempty"`
	Name           string                 `json:"name,omitempty"`
	Priority       int                    `json:"priority,omitempty"`
	ID             int                    `json:"id,omitempty"`
	Comment        string                 `json:"comment,omitempty"`
	Selective      bool                   `json:"selective,omitempty"`
	SecondaryKeys  []string               `json:"secondary_keys,omitempty"`
	Constant       bool                   `json:"constant,omitempty"`
	Position       string                 `json:"position,omitempty"`
}