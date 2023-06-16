package repos

import "time"

type BaseEntity struct {
	PrimitiveId  ID        `json:"_id" bson:"_id" graphql:"ignore"`
	Id           ID        `json:"id" bson:"id"`
	CreationTime time.Time `json:"creationTime"`
	UpdateTime   time.Time `json:"updateTime"`
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
	return c.Id == ""
}
