package entities

type ManagedServicePluginInputField struct {
	Input       string `json:"input"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`

	Unit        *string `json:"unit,omitempty"`
	DisplayUnit *string `json:"displayUnit,omitempty"`

	Type         string   `json:"type"`
	DefaultValue any      `json:"defaultValue,omitempty"`
	Min          *float64 `json:"min,omitempty"`
	Max          *float64 `json:"max,omitempty"`

	Required *bool `json:"required,omitempty"`

	Multiplier *float64 `json:"multiplier,omitempty"`
}

// type ManagedServicePluginSpec struct {
// ApiVersion string         `json:"apiVersion"`
// Services   []ServiceEntry `json:"services"`
// }

type ManagedServicePluginServices struct {
	Kind        string                           `json:"kind"`
	Description string                           `json:"description"`
	Active      bool                             `json:"active"`
	Inputs      []ManagedServicePluginInputField `json:"fields"`
	Resources   []ManagedServicePluginResources  `json:"resources"`
}

type ManagedServicePluginResources struct {
	Kind        string                           `json:"kind"`
	Description string                           `json:"description"`
	Inputs      []ManagedServicePluginInputField `json:"inputs"`
}

type ManagedServicePluginMetadata struct {
	Logo string `json:"logo"`
}

type ManagedServicePlugin struct {
	// Plugin   string                       `json:"plugin"`
	// Metadata ManagedServicePluginMetadata `json:"metadata,omitempty"`
	// Spec     PluginSpec                   `json:"spec"`

	Plugin string `json:"plugin"`
	Meta   struct {
		Logo string `json:"logo"`
	} `json:"meta,omitempty"`
	Spec struct {
		ApiVersion string `json:"apiVersion"`
		Services   []struct {
			Kind        string                           `json:"kind"`
			Description string                           `json:"description"`
			Active      bool                             `json:"active"`
			Inputs      []ManagedServicePluginInputField `json:"inputs"`
			Resources   []struct {
				Kind        string                           `json:"kind"`
				Description string                           `json:"description"`
				Inputs      []ManagedServicePluginInputField `json:"inputs"`
			} `json:"resources"`
		} `json:"services"`
	} `json:"spec"`
}

type ManagedServicePlugins struct {
	Category string                 `json:"category" graphql:"noinput"`
	Items    []ManagedServicePlugin `json:"items" graphql:"noinput"`
}
