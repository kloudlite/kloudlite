package device

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/pkg/constants"
	fn "github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	"github.com/kloudlite/operator/pkg/templates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) rolloutWireguard(req *rApi.Request[*wgv1.Device]) error {
	ctx, obj := req.Context(), req.Object
	depName := fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)
	deployment, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, depName), &appsv1.Deployment{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string, 1)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	if err := r.Update(ctx, deployment); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) rolloutCoreDNS(req *rApi.Request[*wgv1.Device], corefileConfig string) error {
	ctx, obj := req.Context(), req.Object

	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s.%s.svc.cluster.local:17171/resync", obj.Name, obj.Namespace), bytes.NewBuffer([]byte(corefileConfig)))
	if err != nil {
		return err
	}

	logger := r.logger.WithName("rollout-coredns")

	logger.Infof("making request to resync coredns configuration")
	resp, err := http.DefaultClient.Do(hreq)
	if err != nil {
		return err
	}

	logger.Infof("completed request to resync coredns configuration, with status code: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code, received: %d, corefile not synced", resp.StatusCode)
	}

	return nil
}

func (r *Reconciler) generateDeviceConfig(req *rApi.Request[*wgv1.Device]) (devConfig []byte, serverConfig []byte, errr error) {
	ctx, obj := req.Context(), req.Object

	secName := fmt.Sprint(DEVICE_KEY_PREFIX, obj.Name)

	if err := func() error {
		wgService, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, fmt.Sprint(WG_SERVER_NAME_PREFIX, obj.Name)), &corev1.Service{})
		if err != nil {
			return err
		}

		sec, err := rApi.Get(ctx, r.Client, fn.NN(obj.Namespace, secName), &corev1.Secret{})
		if err != nil {
			return err
		}

		pub, priv, ip, err := parseDeviceSec(sec)
		if err != nil {
			return err
		}

		serverPublicKey, sPriv, sIp, err := parseServerSec(sec)
		if err != nil {
			return err
		}

		out, err := templates.Parse(templates.Wireguard.DeviceConfig, deviceConfig{
			DeviceIp:        string(ip),
			DevicePvtKey:    string(priv),
			ServerPublicKey: string(serverPublicKey),
			ServerEndpoint:  fmt.Sprintf("%s:%d", r.Env.DnsHostedZone, wgService.Spec.Ports[0].NodePort),
			DNS:             "10.13.0.3",
			PodCidr:         r.Env.ClusterPodCidr,
			SvcCidr:         r.Env.ClusterServiceCidr,
		})
		if err != nil {
			return err
		}

		// setting devConfig
		devConfig = out

		data := Data{
			ServerIp:         string(sIp) + "/32",
			ServerPrivateKey: string(sPriv),
			DNS:              "127.0.0.1",
			Peers: []Peer{
				{
					PublicKey:  string(pub),
					AllowedIps: string(ip),
				},
			},
		}

		conf, err := templates.Parse(templates.Wireguard.Config, data)
		if err != nil {
			return err
		}

		serverConfig = conf
		return nil
	}(); err != nil {
		return nil, nil, err
	}

	return devConfig, serverConfig, nil
}

func (r *Reconciler) applyDeviceConfig(req *rApi.Request[*wgv1.Device], deviceConfig []byte, serverConfig []byte) error {
	ctx, obj := req.Context(), req.Object
	configName := fmt.Sprint(DEVICE_CONFIG_PREFIX, obj.Name)

	if err := fn.KubectlApply(ctx, r.Client, &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: configName, Namespace: obj.Namespace,
			Labels:          map[string]string{constants.WGDeviceNameKey: obj.Name},
			OwnerReferences: []metav1.OwnerReference{fn.AsOwner(obj, true)},
		},
		Data: map[string][]byte{"config": deviceConfig, "server-config": serverConfig, "sysctl": []byte(`
net.ipv4.ip_forward=1
				`)},
	}); err != nil {
		return err
	}

	if err := r.rolloutWireguard(req); err != nil {
		return err
	}

	return nil
}
