package types

type ActionMeta struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Sample struct {
	DisplayName string     `json:"displayName"`
	CreatedBy   ActionMeta `json:"createdBy"`
	UpdatedBy   ActionMeta `json:"updatedBy"`
	Age         int        `json:"age"`
}
