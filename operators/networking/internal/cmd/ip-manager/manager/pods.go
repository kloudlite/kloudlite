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
	json_patch "github.com/kloudlite/operator/pkg/json-patch"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ct "github.com/kloudlite/operator/apis/common-types"
)

func sanitizePodIP(podIP string) string {
	return strings.ReplaceAll(podIP, ".", "-")
}

const (
	podReservationLabel = "kloudlite.io/podbinding.reservation"
)

var unReservedPodsLabels = map[string]string{
	podReservationLabel: "false",
}

type RegisterPodResult struct {
	PodBindingIP     string
	ReservationToken string
}

func (m *Manager) EnsureIPManagerConfigExists(ctx context.Context) (podIPCounter int, svcIPCounter int, err error) {
	cfgMap, err := m.kclientset.CoreV1().ConfigMaps(m.Env.IPManagerConfigNamespace).Get(ctx, m.Env.IPManagerConfigName, metav1.GetOptions{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return podIPCounter, svcIPCounter, NewError(err, fmt.Sprintf("k8s get ip manager config %s/%s", m.Env.IPManagerConfigNamespace, m.Env.IPManagerConfigName))
		}

		cfgMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.Env.IPManagerConfigName,
				Namespace: m.Env.IPManagerConfigNamespace,
			},
			Data: map[string]string{
				"counter_pod_ip": "0",
				"counter_svc_ip": "10", // reserving 10 svc ips for future usecases
			},
		}

		cfgMap, err = m.kclientset.CoreV1().ConfigMaps(m.Env.IPManagerConfigNamespace).Create(ctx, cfgMap, metav1.CreateOptions{})
		if err != nil {
			return podIPCounter, svcIPCounter, NewError(err, fmt.Sprintf("k8s create ip manager config %s/%s", m.Env.IPManagerConfigNamespace, m.Env.IPManagerConfigName))
		}
	}

	podIPCounter, err = strconv.Atoi(cfgMap.Data["counter_pod_ip"])
	if err != nil {
		return podIPCounter, svcIPCounter, NewError(err, fmt.Sprintf("parsing pod ip counter from configmap %s/%s", m.Env.IPManagerConfigNamespace, m.Env.IPManagerConfigName))
	}
	svcIPCounter, err = strconv.Atoi(cfgMap.Data["counter_svc_ip"])
	if err != nil {
		return podIPCounter, svcIPCounter, NewError(err, fmt.Sprintf("parsing svc ip counter from configmap %s/%s", m.Env.IPManagerConfigNamespace, m.Env.IPManagerConfigName))
	}

	return podIPCounter, svcIPCounter, nil
}

func (m *Manager) CreatePodBindings(ctx context.Context, count int) error {
	podIPCounter, _, err := m.EnsureIPManagerConfigExists(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		s, err := iputils.GenIPAddr(m.Env.ClusterCIDR, m.podIPOffset+podIPCounter+i+1)
		if err != nil {
			return err
		}

		k, err := wgtypes.GenerateKey()
		if err != nil {
			return err
		}

		pb := &networkingv1.PodBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:   sanitizePodIP(s),
				Labels: unReservedPodsLabels,
			},
			Spec: networkingv1.PodBindingSpec{
				GlobalIP:     s,
				WgPrivateKey: k.String(),
				WgPublicKey:  k.PublicKey().String(),
				PodRef:       nil,
				AllowedIPs:   append(strings.Split(m.Env.PodAllowedIPs, ","), m.Env.GatewayInternalDNSNameserver, s),
			},
		}
		pb.EnsureGVK()
		pb.SetAnnotations(pb.GetEnsuredAnnotations())
		if err := m.kcli.Create(ctx, pb); err != nil {
			if !apiErrors.IsAlreadyExists(err) {
				return err
			}
			m.logger.Warn("pod binding already exists", "ip", s)
		}

		m.podPeers[pb.Spec.GlobalIP] = genGatewayWgPodPeer(pb)
	}

	if err := m.RestartWireguard(); err != nil {
		return NewError(err, "while restarting wireguard")
	}

	d := json_patch.Document{}
	d.Add("replace", "/data/counter_pod_ip", strconv.Itoa(podIPCounter+count))
	b, err := d.Json()
	if err != nil {
		return err
	}

	if _, err := m.kclientset.CoreV1().ConfigMaps(m.Env.IPManagerConfigNamespace).Patch(ctx, m.Env.IPManagerConfigName, types.JSONPatchType, b, metav1.PatchOptions{}); err != nil {
		return NewError(err, "patching configmap")
	}

	return nil
}

func (m *Manager) PickFreePodBinding(ctx context.Context) (*networkingv1.PodBinding, error) {
	var pblist networkingv1.PodBindingList
	if err := m.kcli.List(ctx, &pblist, &client.ListOptions{
		Limit:         1,
		LabelSelector: apiLabels.SelectorFromValidatedSet(unReservedPodsLabels),
	}); err != nil {
		return nil, NewError(err, "k8s list pod bindings")
	}

	if len(pblist.Items) == 1 {
		return &pblist.Items[0], nil
	}

	if err := m.CreatePodBindings(ctx, 100); err != nil {
		return nil, NewError(err, "creating pod bindings")
	}

	return m.PickFreePodBinding(ctx)
}

func (m *Manager) RegisterPod(ctx context.Context) (*RegisterPodResult, error) {
	m.Lock()
	defer m.Unlock()

	pb, err := m.PickFreePodBinding(ctx)
	if err != nil {
		return nil, NewError(err, "picking free pod binding")
	}

	reservationToken := fn.CleanerNanoid(40)

	lb := pb.GetLabels()
	if lb == nil {
		lb = make(map[string]string, 1)
	}

	lb[podReservationLabel] = reservationToken
	pb.SetLabels(lb)

	if err := m.kcli.Update(ctx, pb); err != nil {
		return nil, NewError(err, "updating pod binding")
	}

	return &RegisterPodResult{
		PodBindingIP:     pb.Spec.GlobalIP,
		ReservationToken: reservationToken,
	}, nil
}

func (m *Manager) DeregisterPod(ctx context.Context, podBindingIP string, reservationToken string) error {
	m.Lock()
	defer m.Unlock()

	var pb networkingv1.PodBinding
	if err := m.kcli.Get(ctx, fn.NN("", sanitizePodIP(podBindingIP)), &pb); err != nil {
		return NewError(err, "k8s get pod binding")
	}

	if pb.GetLabels()[podReservationLabel] != reservationToken {
		log.Warn("pod reservation token mismatch", "ip", podBindingIP, "reservation-token", reservationToken)
		return nil
	}

	pb.Spec.PodRef = nil
	lb := pb.GetLabels()
	if lb != nil {
		lb[podReservationLabel] = "false"
		pb.SetLabels(lb)
	}
	pb.SetAnnotations(pb.GetEnsuredAnnotations())

	if err := m.kcli.Update(ctx, &pb); err != nil {
		return NewError(err, "k8s update pod binding")
	}

	delete(m.podPeers, pb.Spec.GlobalIP)
	if err := m.RestartWireguardInline(); err != nil {
		return NewError(err, "while restarting wireguard")
	}

	return nil
}

type WgConfigForReservedPodArgs struct {
	ReservationToken string
	PodNamespace     string
	PodName          string
	PodIP            string
}

func (m *Manager) GetWgConfigForReservedPod(ctx context.Context, args WgConfigForReservedPodArgs) ([]byte, error) {
	var pblist networkingv1.PodBindingList
	if err := m.kcli.List(ctx, &pblist, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			podReservationLabel: args.ReservationToken,
		}),
	}); err != nil {
		return nil, NewError(err, "k8s list pod bindings")
	}

	if len(pblist.Items) != 1 {
		return nil, NewError(fmt.Errorf("expected 1 pod binding, got %d", len(pblist.Items)), "k8s list pod bindings")
	}

	pb := pblist.Items[0]

	wgconfig := fmt.Sprintf(`[Interface]
ListenPort = 51820
Address = %s
PrivateKey = %s

DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
`,
		fmt.Sprintf("%s/32", pb.Spec.GlobalIP),
		pb.Spec.WgPrivateKey,

		m.Env.GatewayInternalDNSNameserver,

		m.Env.GatewayWGPublicKey,
		m.Env.GatewayWGEndpoint,
		strings.Join(pb.Spec.AllowedIPs, ", "),
	)

	pb.Spec.PodRef = &ct.NamespacedResourceRef{
		Name:      args.PodName,
		Namespace: args.PodNamespace,
	}
	pb.Spec.PodIP = &args.PodIP
	pb.SetAnnotations(pb.GetEnsuredAnnotations())
	if err := m.kcli.Update(ctx, &pb); err != nil {
		return nil, NewError(err, "updating pod binding")
	}

	m.podPeers[pb.Spec.GlobalIP] = genGatewayWgPodPeer(&pb)
	if err := m.RestartWireguardInline(); err != nil {
		return nil, NewError(err, "while restarting wireguard")
	}
	return []byte(wgconfig), nil
}
