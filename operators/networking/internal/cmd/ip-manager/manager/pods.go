package manager

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/iputils"
	jp "github.com/kloudlite/operator/pkg/json-patch"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

func sanitizePodIP(podIP string) string {
	return strings.ReplaceAll(podIP, ".", "-")
}

type RegisterPodResult struct {
	PodIP  string
	PodUID string
}

func (m *Manager) ensurePodBinding(ctx context.Context, podIP string) (*networkingv1.PodBinding, error) {
	var podBinding networkingv1.PodBinding
	if err := m.kcli.Get(ctx, fn.NN("", sanitizePodIP(podIP)), &podBinding); err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, NewError(err, "k8s get pod binding")
		}

		key, err := wgtypes.GenerateKey()
		if err != nil {
			return nil, Error{Err: err, Message: "while generating wireguard key-pair"}
		}

		podBinding := &networkingv1.PodBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: sanitizePodIP(podIP),
			},
			Spec: networkingv1.PodBindingSpec{
				GlobalIP:     podIP,
				WgPrivateKey: key.String(),
				WgPublicKey:  key.PublicKey().String(),
				AllowedIPs:   strings.Split(m.Env.PodAllowedIPs, ","),
			},
		}

		podBinding.EnsureGVK()
		if err := m.kcli.Create(ctx, podBinding); err != nil {
			return nil, NewError(err, "creating pod binding")
		}

		return podBinding, nil
	}

	return &podBinding, nil
}

func (m *Manager) RegisterPod(ctx context.Context) (*RegisterPodResult, error) {
	m.Lock()
	defer m.Unlock()

	var counterPatch jp.Document

	podIP, isFreeIP, err := func() (string, bool, error) {
		if len(m.FreePodIPs) > 0 {
			ip := m.FreePodIPs[0]
			m.FreePodIPs = m.FreePodIPs[1:]

			counterPatch.Add("add", "/data/free_pod_ips", strings.Join(m.FreePodIPs, ","))
			return ip, true, nil
		}

		counterPatch.Add("add", "/data/counter_pod_ip", strconv.Itoa(m.PodIPCounter+1))
		s, err := iputils.GenIPAddr(m.Env.ClusterCIDR, m.podIPOffset+m.PodIPCounter+1)
		return s, false, err
	}()
	if err != nil {
		return nil, Error{Err: err, Message: "while generating pod IP"}
	}

	pb, err := m.ensurePodBinding(ctx, podIP)
	if err != nil {
		return nil, NewError(err, "while ensuring pod binding")
	}

	b, err := counterPatch.Json()
	if err != nil {
		return nil, NewError(err, "marshalling patches into json array")
	}

	if _, err := m.kclientset.CoreV1().ConfigMaps(m.Env.IPManagerConfigNamespace).Patch(ctx, m.Env.IPManagerConfigName, types.JSONPatchType, b, metav1.PatchOptions{}); err != nil {
		return nil, NewError(err, "patching configmap")
	}

	if !isFreeIP {
		m.PodIPCounter++
	}

	m.podPeers[podIP] = genGatewayWgPodPeer(pb)
	if err := m.RestartWireguard(); err != nil {
		return nil, NewError(err, "while restarting wireguard")
	}

	return &RegisterPodResult{PodIP: podIP, PodUID: string(pb.GetUID())}, nil
}

func (m *Manager) DeregisterPod(ctx context.Context, podBindingIP string, podBindingUID string) error {
	m.Lock()
	defer m.Unlock()

	var podBinding networkingv1.PodBinding
	if err := m.kcli.Get(ctx, fn.NN("", sanitizePodIP(podBindingIP)), &podBinding); err != nil {
		return NewError(err, "k8s get pod binding")
	}

	if string(podBinding.GetUID()) != podBindingUID {
		log.Warn("pod uid mismatch", "ip", podBindingIP, "UID", string(podBinding.GetUID()), "request:UID", podBindingUID)
		return nil
	}

	m.FreePodIPs = append(m.FreePodIPs, podBindingIP)

	var counterPatch jp.Document
	counterPatch.Add("add", "/data/free_pod_ips", strings.Join(m.FreePodIPs, ","))

	b, err := counterPatch.Json()
	if err != nil {
		return NewError(err, "marshalling patches into json array")
	}

	if _, err := m.kclientset.CoreV1().ConfigMaps(m.Env.IPManagerConfigNamespace).Patch(ctx, m.Env.IPManagerConfigName, types.JSONPatchType, b, metav1.PatchOptions{}); err != nil {
		return NewError(err, "patching configmap")
	}

	if err := m.kcli.Delete(ctx, &podBinding); err != nil {
		return NewError(err, "k8s delete pod binding")
	}

	delete(m.podPeers, podBindingIP)
	if err := m.RestartWireguard(); err != nil {
		return NewError(err, "while restarting wireguard")
	}

	return nil
}

func (m *Manager) GetWgConfigForPod(ctx context.Context, podIP string) ([]byte, error) {
	var podBinding networkingv1.PodBinding
	if err := m.kcli.Get(ctx, fn.NN("", sanitizePodIP(podIP)), &podBinding); err != nil {
		return nil, NewError(err, "k8s get pod binding")
	}

	wgconfig := fmt.Sprintf(`
[Interface]
Address = %s
PrivateKey = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
PersistentKeepalive = 5
`,
		fmt.Sprintf("%s/32", podIP),
		podBinding.Spec.WgPrivateKey,

		m.Env.GatewayWGPublicKey,
		m.Env.GatewayWGEndpoint,
		strings.Join(podBinding.Spec.AllowedIPs, ", "),
	)

	return []byte(wgconfig), nil
}
