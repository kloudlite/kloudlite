package entities

type ManagedServiceType string
type ManagedResourceType string

type ManagedServiceTemplate struct {
	Name        string           `json:"name"`
	DisplayName string           `json:"display_name"`
	Fields      []TemplateField  `json:"fields"`
	Output      []TemplateOutput `json:"output"`
}

type TemplateOutput struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type TemplateField struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Description  string `json:"description"`
	Min          int    `json:"min"`
	Max          int    `json:"max"`
	DefaultValue string `json:"default_value"`
	Hidden       bool   `json:"hidden"`
	InputType    string `json:"input_type"`
	Unit         string `json:"unit"`
	Required     bool   `json:"required"`
}

type ManagedResourceTemplate struct {
	Name        string           `json:"name"`
	DisplayName string           `json:"display_name"`
	Fields      []*TemplateField `json:"fields"`
}
