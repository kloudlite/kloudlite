package workmachinesessions

type WMSessionManager struct {
	secret string
}

func NewWMSessionManager(secret string) *WMSessionManager {
	return &WMSessionManager{
		secret: secret,
	}
}

func (wm *WMSessionManager) CreateWMSession(machineName string, cluster string, user string) {

}

func (wm *WMSessionManager) ValidateWMSession(machineName string, cluster string, user string) {

}
