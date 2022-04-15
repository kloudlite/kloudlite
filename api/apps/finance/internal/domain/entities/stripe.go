package entities

// import "kloudlite.io/pkg/repos"

type StripeConstants string

const (
	DEV_STRIPE_CUSTOMER_ID   = StripeConstants("dev-customer-id")
	ROBOT_STRIPE_INTENT_ID   = StripeConstants("robot-intent-id")
	ROBOT_STRIPE_CUSTOMER_ID = StripeConstants("robot-customer-id")
)

// type Stripe struct {
// 	repos.BaseEntity `bson:",inline"`
// 	Index            int          `json:"index" bson:"index"`
// 	Name             string       `json:"name" bson:"name"`
// 	ClusterId        repos.ID     `json:"cluster_id" bson:"cluster_id"`
// 	UserId           repos.ID     `json:"user_id" bson:"user_id"`
// 	PrivateKey       *string      `json:"private_key" bson:"private_key"`
// 	PublicKey        *string      `json:"public_key" bson:"public_key"`
// 	Ip               string       `json:"ip" bson:"ip"`
// 	Status           DeviceStatus `json:"status" bson:"status"`
// }
