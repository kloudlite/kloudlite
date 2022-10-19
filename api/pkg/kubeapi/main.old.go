package kubeapi

// type Client struct {
// 	Address string `env:"KUBE_API_ADDRESS"`
// }
//
// func (c *Client) GetServiceIp(ctx context.Context, namespace, name string) (string, error) {
// 	service := v1.Service{}
// 	get, err := http.Get(c.Address + "/api/v1/namespaces/" + namespace + "/services/" + name)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer get.Body.Close()
// 	all, err := io.ReadAll(get.Body)
// 	if err != nil {
// 		return "", err
// 	}
// 	if err := json.Unmarshal(all, &service); err != nil {
// 		return "", err
// 	}
// 	return service.Spec.ClusterIP, nil
// }
//
// func (c *Client) GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
// 	secret := v1.Secret{}
// 	get, err := http.Get(c.Address + "/api/v1/namespaces/" + namespace + "/secrets/" + name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer get.Body.Close()
// 	all, err := io.ReadAll(get.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := json.Unmarshal(all, &secret); err != nil {
// 		return nil, err
// 	}
// 	return &secret, nil
// }
//
// func NewClient(addr string) *Client {
// 	return &Client{
// 		Address: addr,
// 	}
// }
