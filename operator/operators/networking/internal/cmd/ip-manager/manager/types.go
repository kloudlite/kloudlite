package manager

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/networking/internal/cmd/ip-manager/env"
	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/kubectl"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Manager struct {
	kclientset *kubernetes.Clientset
	kcli       client.Client
	Env        *env.Env

	// private fields
	sync.Mutex

	logger *slog.Logger

	podIPOffset int

	PodIPCounter int      `json:"counter_pod_ip"`
	FreePodIPs   []string `json:"free_pod_ips"`

	SvcIPCounter int      `json:"counter_svc_ip"`
	FreeSvcIPs   []string `json:"free_svc_ips"`

	podPeers        map[string]string
	svcNginxStreams map[string][]string

	runningNginxStreamsMD5     string
	runningNginxStreamFileSize int

	// svcBindingsMap      map[string]*ReserveServiceResponse
	gatewayWgExtraPeers string
}

func NewManager(ev *env.Env, logger *slog.Logger, kclientset *kubernetes.Clientset, kcli client.Client) (*Manager, error) {
	cfg := &corev1.ConfigMap{}

	ctx := context.TODO()
	cfg, err := kclientset.CoreV1().ConfigMaps(ev.IPManagerConfigNamespace).Get(ctx, ev.IPManagerConfigName, metav1.GetOptions{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}

		cfg = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: ev.IPManagerConfigName,
			},
			Data: map[string]string{
				"counter_pod_ip": "0",
				"counter_svc_ip": "0",
			},
		}
		cfg, err = kclientset.CoreV1().ConfigMaps(ev.IPManagerConfigNamespace).Create(ctx, cfg, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}

	s := strings.SplitN(ev.ServiceCIDR, "/", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf("invalid service CIDR: %s", ev.ServiceCIDR)
	}

	cidrSuffix, err := strconv.Atoi(s[1])
	if err != nil {
		return nil, err
	}

	manager := Manager{
		Env:         ev,
		Mutex:       sync.Mutex{},
		podIPOffset: int(math.Pow(2, float64(32-cidrSuffix)) + 1),
		kclientset:  kclientset,
		kcli:        kcli,
		podPeers:    make(map[string]string),
		logger:      logger,
		// svcBindingsMap: make(map[string]*ReserveServiceResponse),
	}

	if v, ok := cfg.Data["counter_pod_ip"]; ok {
		manager.PodIPCounter, err = strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
	}

	if v, ok := cfg.Data["counter_svc_ip"]; ok {
		manager.SvcIPCounter, err = strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
	}

	if v, ok := cfg.Data["free_pod_ips"]; ok {
		manager.FreePodIPs = strings.Split(v, ",")
	}

	if v, ok := cfg.Data["free_svc_ips"]; ok {
		manager.FreeSvcIPs = strings.Split(v, ",")
	}

	t := TimerStart(manager.logger, "fetching pod bindings")
	c, err := kubectl.PaginatedList[*networkingv1.PodBinding](ctx, kcli, &networkingv1.PodBindingList{}, &client.ListOptions{
		Limit: 50,
	})
	if err != nil {
		return nil, err
	}
	t.Stop()

	t.Reset("generating pod wg peers")
	for podBinding := range c {
		manager.podPeers[podBinding.Spec.GlobalIP] = genGatewayWgPodPeer(podBinding)
	}
	t.Stop()

	t.Reset("fetching service bindings")
	sbList, err := kubectl.PaginatedList[*networkingv1.ServiceBinding](ctx, kcli, &networkingv1.ServiceBindingList{}, &client.ListOptions{Limit: 10})
	if err != nil {
		return nil, err
	}
	t.Stop()

	t.Reset("generating service binding nginx streams")
	manager.svcNginxStreams = make(map[string][]string)
	for svcBinding := range sbList {
		if svcBinding.Spec.ServiceRef == nil {
			continue
		}
		// manager.svcBindingsMap[fmt.Sprintf("%s/%s", svcBinding.Spec.ServiceRef.Namespace, svcBinding.Spec.ServiceRef.Name)] = &ReserveServiceResponse{
		// 	ServiceBindingIP: svcBinding.Spec.GlobalIP,
		// 	ReservationToken: svcBinding.GetLabels()[svcReservationLabel],
		// }
		manager.svcNginxStreams[svcBinding.Spec.GlobalIP] = RegisterNginxStreamConfig(svcBinding)
	}
	t.Stop()

	if ev.ExtraWireguardPeersPath != "" {
		f, err := os.Open(ev.ExtraWireguardPeersPath)
		if err != nil {
			return nil, errors.NewEf(err, "failed to open extra wireguard peers file %s", ev.ExtraWireguardPeersPath)
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return nil, errors.NewEf(err, "failed to read extra wireguard peers file %s", ev.ExtraWireguardPeersPath)
		}
		manager.gatewayWgExtraPeers = string(b)
	}

	t.Reset("starting wireguard")
	if err := manager.RestartWireguard(); err != nil {
		return nil, err
	}
	t.Stop()

	t.Reset("syncing nginx streams")
	if err := manager.SyncNginxStreams(); err != nil {
		return nil, err
	}
	t.Stop()

	manager.logger.Info("manager initialized with", "counter_pod_ip", manager.PodIPCounter, "counter_svc_ip", manager.SvcIPCounter)

	return &manager, nil
}

type Error struct {
	Err     error
	Message string
	Code    int
}

func (e Error) Error() string {
	msg := e.Message
	if e.Code != 0 {
		msg = fmt.Sprintf("[%d] %s", e.Code, msg)
	}
	if e.Err != nil {
		return errors.NewEf(e.Err, e.Message).Error()
	}
	return msg
}

func NewError(err error, msg string) Error {
	return Error{Err: err, Message: msg}
}

type Timed struct {
	msg    string
	start  time.Time
	logger *slog.Logger
}

func TimerStart(logger *slog.Logger, msg string) *Timed {
	t := time.Now()
	logger.Debug(msg, "at", t)
	return &Timed{
		msg:    msg,
		start:  t,
		logger: logger,
	}
}

func (t *Timed) Stop() {
	t.logger.Info(t.msg, "took", fmt.Sprintf("%.2fs", time.Since(t.start).Seconds()))
}

func (t *Timed) Reset(msg string) {
	t.start = time.Now()
	t.msg = msg
}
