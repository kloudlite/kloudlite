package manager

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
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

	logger *log.Logger

	podIPOffset int

	PodIPCounter int      `json:"counter_pod_ip"`
	FreePodIPs   []string `json:"free_pod_ips"`

	SvcIPCounter int      `json:"counter_svc_ip"`
	FreeSvcIPs   []string `json:"free_svc_ips"`

	podPeers        map[string]string
	svcNginxStreams map[string][]string
}

func NewManager(ev *env.Env, kclientset *kubernetes.Clientset, kcli client.Client) (*Manager, error) {
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
		logger: log.NewWithOptions(os.Stderr, log.Options{
			Level:           log.DebugLevel,
			Prefix:          "[ip-manager] ",
			ReportTimestamp: false,
			ReportCaller:    true,
		}),
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

	c, err := kubectl.PaginatedList[*networkingv1.PodBinding](ctx, kcli, &networkingv1.PodBindingList{}, &client.ListOptions{Limit: 10})
	if err != nil {
		return nil, err
	}

	for podBinding := range c {
		manager.podPeers[podBinding.Spec.GlobalIP] = genGatewayWgPodPeer(podBinding)
	}

	sbList, err := kubectl.PaginatedList[*networkingv1.ServiceBinding](ctx, kcli, &networkingv1.ServiceBindingList{}, &client.ListOptions{Limit: 10})
	if err != nil {
		return nil, err
	}

	manager.svcNginxStreams = make(map[string][]string)
	for svcBinding := range sbList {
		manager.svcNginxStreams[svcBinding.Spec.GlobalIP] = RegisterNginxStreamConfig(svcBinding)
	}

	if err := manager.RestartWireguard(); err != nil {
		return nil, err
	}

	if err := manager.SyncNginxStreams(); err != nil {
		return nil, err
	}

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
