package manager

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
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
	fn "github.com/kloudlite/operator/pkg/functions"
)

func sanitizePodIP(podIP string) string {
	return strings.ReplaceAll(podIP, ".", "-")
}

const (
	podReservationLabel     = "kloudlite.io/podbinding.reservation"
	podNetworkingAllowedIPs = "kloudlite.io/networking.pod.allowedIPs"
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

type RequestContext struct {
	context.Context
	logger *slog.Logger
}

func (m *Manager) CreatePodBindings(ctx RequestContext, count int) error {
	podIPCounter, _, err := m.EnsureIPManagerConfigExists(ctx)
	ctx.logger.Info("creating pod bindings", "pod-ip-counter", podIPCounter, "creation-count", count)
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

		ctx.logger.Info("creating pod binding", "ip", s, "ip-offset", m.podIPOffset+podIPCounter+i+1)
		pb := &networkingv1.PodBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sanitizePodIP(s),
				Namespace: m.Env.IPManagerConfigNamespace,
				Labels:    unReservedPodsLabels,
			},
			Spec: networkingv1.PodBindingSpec{
				GlobalIP:     s,
				WgPrivateKey: k.String(),
				WgPublicKey:  k.PublicKey().String(),
				PodRef:       nil,
				AllowedIPs: append(
					strings.Split(m.Env.PodAllowedIPs, ","),
					m.Env.GatewayInternalDNSNameserver,
					s,
				),
			},
		}
		pb.EnsureGVK()
		pb.SetAnnotations(pb.GetEnsuredAnnotations())
		t := TimerStart(ctx.logger, "creating pod binding")
		if err := m.kcli.Create(ctx, pb); err != nil {
			if !apiErrors.IsAlreadyExists(err) {
				return err
			}
			m.logger.Warn("pod binding already exists", "ip", s)
			<-time.After(500 * time.Millisecond)
			continue
		}
		t.Stop()

		m.podPeers[pb.Spec.GlobalIP] = genGatewayWgPodPeer(pb)
	}

	t := TimerStart(ctx.logger, "restarting wireguard")
	if err := m.RestartWireguardInline(); err != nil {
		ctx.logger.Error("while restarting wireguard", "err", err)
	}
	t.Stop()

	d := json_patch.Document{}
	d.Add("replace", "/data/counter_pod_ip", strconv.Itoa(podIPCounter+count))
	b, err := d.Json()
	if err != nil {
		return err
	}

	ctx2, cf := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cf()

	t.Reset("patching configmap")
	if _, err := m.kclientset.CoreV1().ConfigMaps(m.Env.IPManagerConfigNamespace).Patch(ctx2, m.Env.IPManagerConfigName, types.JSONPatchType, b, metav1.PatchOptions{}); err != nil {
		return NewError(err, "patching configmap")
	}
	t.Stop()

	return nil
}

func (m *Manager) PickFreePodBinding(ctx context.Context, createIfNone bool) (*networkingv1.PodBinding, error) {
	var pblist networkingv1.PodBindingList
	if err := m.kcli.List(ctx, &pblist, &client.ListOptions{
		Limit:         1,
		LabelSelector: apiLabels.SelectorFromValidatedSet(unReservedPodsLabels),
	}); err != nil {
		return nil, NewError(err, "k8s list pod bindings")
	}

	if len(pblist.Items) >= 1 {
		return &pblist.Items[0], nil
	}

	go func() {
		c2, cf := context.WithTimeout(context.TODO(), 30*time.Second)
		defer cf()
		if err := m.CreatePodBindings(RequestContext{c2, m.logger}, 15); err != nil {
			m.logger.Error("creating pod bindings", "err", err)
		}
	}()

	<-time.After(100 * time.Millisecond)
	return m.PickFreePodBinding(ctx, false)
}

func (m *Manager) DeregisterPod(ctx context.Context, podNamespace, podName string) error {
	m.Lock()
	defer m.Unlock()

	reservationToken := fmt.Sprintf("%s.%s", podNamespace, podName)

	var pblist networkingv1.PodBindingList
	if err := m.kcli.List(ctx, &pblist, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			podReservationLabel: reservationToken,
		}),
	}); err != nil {
		return NewError(err, "k8s get pod binding")
	}

	if len(pblist.Items) == 0 {
		m.logger.Info("pod binding not found", "reservation-token", reservationToken)
		return nil
	}

	pb := pblist.Items[0]

	pb.Spec.PodRef = nil
	lb := pb.GetLabels()
	if lb != nil {
		lb[podReservationLabel] = "false"
		pb.SetLabels(lb)
	}
	pb.SetAnnotations(pb.GetEnsuredAnnotations())

	// smp := client.MergeFrom(&pb)
	if err := m.kcli.Update(ctx, &pb); err != nil {
		return NewError(err, "k8s update pod binding")
	}

	// if err := m.kcli.Patch(ctx, &pb, smp); err != nil {
	// 	return NewError(err, "k8s update pod binding")
	// }
	return nil
}

type WgConfigForReservedPodArgs struct {
	PodNamespace string
	PodName      string
	PodIP        string
}

func (m *Manager) GetWgConfigForReservedPod(ctx context.Context, args WgConfigForReservedPodArgs) ([]byte, error) {
	reservationToken := fmt.Sprintf("%s.%s", args.PodNamespace, args.PodName)
	logger := m.logger.With("pod", reservationToken)

	m.Lock()
	defer m.Unlock()

	t := TimerStart(logger, "picking pod-binding")
	var pblist networkingv1.PodBindingList
	if err := m.kcli.List(ctx, &pblist, &client.ListOptions{
		Limit:         1,
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{podReservationLabel: reservationToken}),
	}); err != nil {
		return nil, NewError(err, "k8s list pod bindings")
	}

	var pb *networkingv1.PodBinding
	var err error

	if len(pblist.Items) == 1 {
		pb = &pblist.Items[0]
	} else {
		pb, err = m.PickFreePodBinding(ctx, true)
		if err != nil {
			return nil, NewError(err, "picking free pod binding")
		}
	}

	t.Stop()

	allowedIPs := pb.Spec.AllowedIPs

	var pod corev1.Pod
	if err := m.kcli.Get(ctx, fn.NN(args.PodNamespace, args.PodName), &pod); err != nil {
		return nil, NewError(err, fmt.Sprintf("failed to find pod (%s/%s)", args.PodNamespace, args.PodName))
	}

	if v, ok := pod.GetAnnotations()[podNetworkingAllowedIPs]; ok {
		for _, item := range strings.Split(v, ",") {
			allowedIPs = append(allowedIPs, strings.TrimSpace(item))
		}
	}

	wgconfig := fmt.Sprintf(`[Interface]
ListenPort = 51820
Address = %s
PrivateKey = %s

DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
PersistentKeepalive = 25
`,
		fmt.Sprintf("%s/32", pb.Spec.GlobalIP),
		pb.Spec.WgPrivateKey,

		m.Env.GatewayInternalDNSNameserver,

		m.Env.GatewayWGPublicKey,
		m.Env.GatewayWGEndpoint,
		strings.Join(allowedIPs, ", "),
	)

	pb.Spec.PodRef = &ct.NamespacedResourceRef{
		Name:      args.PodName,
		Namespace: args.PodNamespace,
	}
	pb.Spec.PodIP = &args.PodIP
	lb := pb.GetLabels()
	if lb == nil {
		lb = make(map[string]string, 1)
	}
	lb[podReservationLabel] = reservationToken
	pb.SetLabels(lb)

	pb.SetAnnotations(pb.GetEnsuredAnnotations())

	if err := m.kcli.Update(ctx, pb); err != nil {
		return nil, NewError(err, "updating pod binding")
	}

	m.podPeers[pb.Spec.GlobalIP] = genGatewayWgPodPeer(pb)
	return []byte(wgconfig), nil
}
