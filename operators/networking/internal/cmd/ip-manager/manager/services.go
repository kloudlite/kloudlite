package manager

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ct "github.com/kloudlite/operator/apis/common-types"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/iputils"
	json_patch "github.com/kloudlite/operator/pkg/json-patch"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
)

func sanitizeSvcIP(svcIP string) string {
	return strings.ReplaceAll(svcIP, ".", "-")
}

const (
	svcReservationLabel = "kloudlite.io/servicebinding.reservation"
)

func svcRegistrationValue(namespace, name string) string {
	return fn.Md5([]byte(fmt.Sprintf("%s/%s", namespace, name)))
}

var unReservedServiceLabels = map[string]string{
	svcReservationLabel: "false",
}

func (m *Manager) CreateSvcBindings(ctx context.Context, count int) error {
	_, svcIPCounter, err := m.EnsureIPManagerConfigExists(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		s, err := iputils.GenIPAddr(m.Env.ClusterCIDR, svcIPCounter+i+1)
		if err != nil {
			return err
		}

		sb := &networkingv1.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{Name: sanitizeSvcIP(s), Labels: unReservedServiceLabels},
			Spec: networkingv1.ServiceBindingSpec{
				GlobalIP:   s,
				ServiceIP:  nil,
				ServiceRef: nil,
				Ports:      nil,
			},
		}
		sb.EnsureGVK()
		sb.SetAnnotations(sb.GetEnsuredAnnotations())
		if err := m.kcli.Create(ctx, sb); err != nil {
			if !apiErrors.IsAlreadyExists(err) {
				return err
			}
			m.logger.Warn("svc binding already exists", "ip", s)
		}

		if err := m.WgAddAddr(s); err != nil {
			return err
		}
	}

	d := json_patch.Document{}
	d.Add("replace", "/data/counter_svc_ip", strconv.Itoa(svcIPCounter+count))
	b, err := d.Json()
	if err != nil {
		return err
	}

	if _, err := m.kclientset.CoreV1().ConfigMaps(m.Env.IPManagerConfigNamespace).Patch(ctx, m.Env.IPManagerConfigName, types.JSONPatchType, b, metav1.PatchOptions{}); err != nil {
		return NewError(err, "patching configmap")
	}

	return nil
}

func (m *Manager) PickFreeSvcBinding(ctx context.Context) (*networkingv1.ServiceBinding, error) {
	var sblist networkingv1.ServiceBindingList
	if err := m.kcli.List(ctx, &sblist, &client.ListOptions{
		Limit:         50,
		LabelSelector: apiLabels.SelectorFromValidatedSet(unReservedServiceLabels),
	}); err != nil {
		return nil, NewError(err, "k8s list service bindings")
	}

	if len(sblist.Items) == 50 {
		return &sblist.Items[0], nil
	}

	if err := m.CreateSvcBindings(ctx, 100); err != nil {
		return nil, NewError(err, "creating service bindings")
	}

	return m.PickFreeSvcBinding(ctx)
}

type ReserveServiceResponse struct {
	ServiceBindingIP string `json:"service_binding_ip"`
}

func (m *Manager) ReserveService(ctx context.Context, namespace, name string) (*ReserveServiceResponse, error) {
	m.Lock()
	defer m.Unlock()

	sb, err := m.PickFreeSvcBinding(ctx)
	if err != nil {
		return nil, NewError(err, "picking free service binding")
	}

	lb := sb.Labels
	if lb == nil {
		lb = make(map[string]string, 1)
	}

	lb[svcReservationLabel] = svcRegistrationValue(namespace, name)
	sb.SetLabels(lb)

	sb.Spec.ServiceRef = &ct.NamespacedResourceRef{
		Name:      name,
		Namespace: namespace,
	}
	sb.EnsureGVK()
	sb.SetAnnotations(sb.GetEnsuredAnnotations())
	if err := m.kcli.Update(ctx, sb); err != nil {
		return nil, NewError(err, "updating service binding")
	}

	return &ReserveServiceResponse{
		ServiceBindingIP: sb.Spec.GlobalIP,
	}, nil
}

func (m *Manager) RegisterService(ctx context.Context, svcBindingIP string) error {
	m.Lock()
	defer m.Unlock()

	var sb networkingv1.ServiceBinding
	if err := m.kcli.Get(ctx, fn.NN("", sanitizeSvcIP(svcBindingIP)), &sb); err != nil {
		return NewError(err, "k8s get service binding")
	}

	m.svcNginxStreams[sb.Spec.GlobalIP] = RegisterNginxStreamConfig(&sb)

	return m.SyncNginxStreams()
}

func (m *Manager) DeregisterService(ctx context.Context, svcBindingIP string) error {
	var svcBinding networkingv1.ServiceBinding
	if err := m.kcli.Get(ctx, fn.NN("", sanitizeSvcIP(svcBindingIP)), &svcBinding); err != nil {
		if apiErrors.IsNotFound(err) {
			m.logger.Info("service binding not found, already deleted", "svc_binding_ip", svcBindingIP)
			return nil
		}
		return NewError(err, "k8s get svc binding")
	}

	svcBinding.Spec.ServiceIP = nil
	svcBinding.Spec.ServiceRef = nil
	svcBinding.Spec.Ports = nil

	lb := svcBinding.Labels
	delete(svcBinding.GetLabels(), "kloudlite.io/global.hostname")

	if lb == nil {
		lb = make(map[string]string, 1)
	}
	lb[svcReservationLabel] = "false"
	svcBinding.SetLabels(lb)

	svcBinding.SetAnnotations(svcBinding.GetEnsuredAnnotations())

	if err := m.kcli.Update(ctx, &svcBinding); err != nil {
		return NewError(err, "updating service binding")
	}

	if err := m.WgRemoveAddr(svcBindingIP); err != nil {
		return NewError(err, "removing ip addr from wg")
	}

	delete(m.svcNginxStreams, svcBindingIP)
	return m.SyncNginxStreams()
}
