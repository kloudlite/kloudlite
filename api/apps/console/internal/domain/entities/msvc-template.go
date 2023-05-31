package entities

type InputField struct {
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	InputType    string   `json:"inputType"`
	DefaultValue any      `json:"defaultValue"`
	Min          *float64 `json:"min,omitempty"`
	Max          *float64 `json:"max,omitempty"`
	Required     *bool    `json:"required,omitempty"`
	Unit         *string  `json:"unit,omitempty"`
}

type OutputField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type MsvcTemplateEntry struct {
	Name            string         `json:"name"`
	LogoUrl         string         `json:"logoUrl"`
	DisplayName     string         `json:"displayName"`
	Description     string         `json:"description"`
	Active          bool           `json:"active"`
	Fields          []InputField   `json:"fields"`
	// InputMiddleware *string        `json:"inputMiddleware"`
	Outputs         []OutputField  `json:"outputs"`
	Resources       []MresTemplate `json:"resources"`
}

type MresTemplate struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"displayName"`
	Description string        `json:"description"`
	Fields      []InputField  `json:"fields"`
	Outputs     []OutputField `json:"outputs"`
}

type MsvcTemplate struct {
	Category    string      `json:"category"`
	DisplayName string      `json:"displayName"`
	Items       []MsvcTemplateEntry `json:"items"`
}
