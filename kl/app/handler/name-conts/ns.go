package ns

type Action string
type ItemName string

const (
	AuthBtn         ItemName = "AuthBtn"
	AccountTitle    ItemName = "AccountTitle"
	DeviceBtn       ItemName = "DeviceBtn"
	UpdateBtn       ItemName = "UpdateBtn"
	UserBtn         ItemName = "UserBtn"
	AccountItem     ItemName = "AccountItem"
	AccountBtn      ItemName = "AccountBtn"
	AccountSettings ItemName = "AccountSettings"

	EnvTitle ItemName = "EnvTitle"

	EnvBtn ItemName = "EnvBtn"
)

const (
	Logout              Action = "Logout"
	Login               Action = "Login"
	ToggleDevice        Action = "ToggleDevice"
	UpdateClient        Action = "UpdateClient"
	SwitchAccount       Action = "SwitchAccount"
	OpenAccountSettings Action = "OpenAccountSettings"
)
