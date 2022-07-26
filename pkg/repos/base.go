package repos

import "time"

type BaseEntity struct {
	Id           ID        `json:"id" bson:"id"`
	CreationTime time.Time `json:"creation_time" bson:"creation_time"`
	UpdateTime   time.Time `json:"update_time" bson:"update_time"`
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
