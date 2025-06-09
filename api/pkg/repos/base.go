package repos

import (
	"time"
)

type BaseEntity struct {
	PrimitiveId       ID        `json:"_id,omitempty" graphql:"ignore" struct-json-path:",ignore"`
	Id                ID        `json:"id" graphql:"scalar-type=ID"`
	CreationTime      time.Time `json:"creationTime"`
	UpdateTime        time.Time `json:"updateTime"`
	RecordVersion     int       `json:"recordVersion"`
	MarkedForDeletion *bool     `json:"markedForDeletion,omitempty"`
}

func (c *BaseEntity) GetPrimitiveID() ID {
	return c.PrimitiveId
}

func (c *BaseEntity) GetId() ID {
	return c.Id
}

func (c *BaseEntity) GetUpdateTime() time.Time {
	return c.UpdateTime
}

func (c *BaseEntity) GetCreationTime() time.Time {
	return c.CreationTime
}

func (c *BaseEntity) SetId(id ID) {
	c.Id = id
}

func (c *BaseEntity) SetCreationTime(ct time.Time) {
	c.CreationTime = ct
}

func (c *BaseEntity) SetUpdateTime(ut time.Time) {
	c.UpdateTime = ut
}

func (c *BaseEntity) IsZero() bool {
	return c == nil || c.Id == ""
}

func (c *BaseEntity) IncrementRecordVersion() {
	c.RecordVersion += 1
}

func (c *BaseEntity) GetRecordVersion() int {
	return c.RecordVersion
}

func (c *BaseEntity) IsMarkedForDeletion() bool {
	if c.MarkedForDeletion == nil {
		return false
	}
	return *c.MarkedForDeletion
}

// added because gqlgen needs it when using @key directives
// read more [here](https://github.com/99designs/gqlgen/blob/ee526b05f28b0e7d5a8e7b1da28da3e03c826df9/plugin/federation/fedruntime/runtime.go#L12)
func (c *BaseEntity) IsEntity() {
}
