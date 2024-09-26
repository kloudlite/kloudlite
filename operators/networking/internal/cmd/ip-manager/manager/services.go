package manager

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ct "github.com/kloudlite/operator/apis/common-types"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/pkg/iputils"
	json_patch "github.com/kloudlite/operator/pkg/json-patch"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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

var unReservedServiceLabels = map[string]string{
	svcReservationLabel: "false",
}

func (m *Manager) CreateSvcBindings(ctx context.Context, count int) error {
	m.logger.Info("Creating service bindings", "count", count)
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
			ObjectMeta: metav1.ObjectMeta{Name: sanitizeSvcIP(s), Namespace: m.Env.IPManagerConfigNamespace, Labels: unReservedServiceLabels},
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
		Limit:         10,
		LabelSelector: apiLabels.SelectorFromValidatedSet(unReservedServiceLabels),
	}); err != nil {
		return nil, NewError(err, "k8s list service bindings")
	}

	if len(sblist.Items) == 10 {
		return &sblist.Items[0], nil
	}

	if err := m.CreateSvcBindings(ctx, 15); err != nil {
		return nil, NewError(err, "creating service bindings")
	}

	return m.PickFreeSvcBinding(ctx)
}

func (m *Manager) ReserveService(ctx context.Context, namespace, name string, ports []corev1.ServicePort) error {
	m.Lock()
	defer m.Unlock()

	reservationToken := fmt.Sprintf("%s.%s", namespace, name)

	sb, err := func() (*networkingv1.ServiceBinding, error) {
		sb2, err2 := m.getServiceBinding(ctx, namespace, name)
		if err2 != nil {
			if errors.Is(err2, ErrServiceBindingNotFound) {
				return m.PickFreeSvcBinding(ctx)
			}
			return nil, NewError(err2, "k8s get service binding")
		}
		return sb2, nil
	}()
	if err != nil {
		return NewError(err, "picking free service binding")
	}

	lb := sb.Labels
	if lb == nil {
		lb = make(map[string]string, 1)
	}
	lb[svcReservationLabel] = reservationToken
	sb.SetLabels(lb)

	sb.Spec.ServiceRef = &ct.NamespacedResourceRef{
		Name:      name,
		Namespace: namespace,
	}
	sb.Spec.Ports = ports

	sb.EnsureGVK()
	sb.SetAnnotations(sb.GetEnsuredAnnotations())
	if err := m.kcli.Update(ctx, sb); err != nil {
		return NewError(err, "updating service binding")
	}

	if err := m.WgAddAddr(sb.Spec.GlobalIP); err != nil {
		return NewError(err, "adding ip addr from wg")
	}

	m.svcNginxStreams[sb.Spec.GlobalIP] = RegisterNginxStreamConfig(sb)
	if err := m.SyncNginxStreams(); err != nil {
		return NewError(err, "syncing nginx streams")
	}
	m.logger.Info("nginx successfully synced")

	return nil
}

var ErrServiceBindingNotFound = fmt.Errorf("servicebinding not found")

func (m *Manager) getServiceBinding(ctx context.Context, namespace, name string) (*networkingv1.ServiceBinding, error) {
	var sblist networkingv1.ServiceBindingList
	if err := m.kcli.List(ctx, &sblist, &client.ListOptions{
		Limit: 1,
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			svcReservationLabel: fmt.Sprintf("%s.%s", namespace, name),
		}),
	}); err != nil {
		return nil, NewError(err, "k8s list service bindings")
	}

	if len(sblist.Items) == 0 {
		return nil, ErrServiceBindingNotFound
	}

	if len(sblist.Items) != 1 {
		return nil, fmt.Errorf("expected 1 service binding, got %d", len(sblist.Items))
	}

	return &sblist.Items[0], nil
}

func (m *Manager) DeregisterService(ctx context.Context, namespace, name string) error {
	sb, err := m.getServiceBinding(ctx, namespace, name)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			m.logger.Info("service binding not found, already deleted", "svc", fmt.Sprintf("%s/%s", namespace, name))
			return nil
		}
		return err
	}

	sb.Spec.ServiceIP = nil
	sb.Spec.ServiceRef = nil
	sb.Spec.Ports = nil
	sb.Spec.Hostname = ""

	sb.SetLabels(map[string]string{
		svcReservationLabel: "false",
	})

	delete(sb.Annotations, "kloudlite.io/global.hostname")
	sb.SetAnnotations(sb.GetEnsuredAnnotations())

	if err := m.kcli.Update(ctx, sb); err != nil {
		return NewError(err, "updating service binding")
	}

	if err := m.WgRemoveAddr(sb.Spec.GlobalIP); err != nil {
		return NewError(err, "removing ip addr from wg")
	}

	delete(m.svcNginxStreams, sb.Spec.GlobalIP)
	return m.SyncNginxStreams()
}
