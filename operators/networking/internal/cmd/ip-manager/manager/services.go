package manager

import (
	"context"
	"strconv"
	"strings"

	ct "github.com/kloudlite/operator/apis/common-types"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	"github.com/kloudlite/operator/pkg/iputils"
	jp "github.com/kloudlite/operator/pkg/json-patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

func sanitizeSvcIP(svcIP string) string {
	return strings.ReplaceAll(svcIP, ".", "-")
}

type RegisterServiceResponse struct {
	ServiceBindingIP  string `json:"service_binding_ip"`
	ServiceBindingUID string `json:"service_binding_uid"`
}

func (m *Manager) RegisterService(ctx context.Context, namespace, name string) (*RegisterServiceResponse, error) {
	m.Lock()
	defer m.Unlock()

	var counterPatch jp.Document

	svcIP, isFreeIP, err := func() (string, bool, error) {
		if len(m.FreeSvcIPs) > 0 {
			ip := m.FreeSvcIPs[0]
			m.FreeSvcIPs = m.FreeSvcIPs[1:]

			counterPatch.Add("add", "/data/free_svc_ips", strings.Join(m.FreeSvcIPs, ","))
			return ip, true, nil
		}

		counterPatch.Add("add", "/data/counter_svc_ip", strconv.Itoa(m.SvcIPCounter+1))
		s, err := iputils.GenIPAddr(m.Env.ServiceCIDR, m.SvcIPCounter+1)
		return s, false, err
	}()
	if err != nil {
		return nil, Error{Err: err, Message: "while generating svc IP"}
	}

	svcBinding := &networkingv1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: sanitizeSvcIP(svcIP),
		},
		Spec: networkingv1.ServiceBindingSpec{
			GlobalIP:   svcIP,
			ServiceIP:  new(string),
			ServiceRef: ct.NamespacedResourceRef{Name: name, Namespace: namespace},
		},
	}

	svcBinding.EnsureGVK()
	if err := m.kcli.Create(ctx, svcBinding); err != nil {
		return nil, NewError(err, "creating svc binding")
	}

	b, err := counterPatch.Json()
	if err != nil {
		return nil, NewError(err, "marshalling patches into json array")
	}

	if _, err := m.kclientset.CoreV1().ConfigMaps(m.Env.IPManagerConfigNamespace).Patch(ctx, m.Env.IPManagerConfigName, types.JSONPatchType, b, metav1.PatchOptions{}); err != nil {
		return nil, NewError(err, "patching configmap")
	}

	// if err := m.WgAddAddr(svcIP); err != nil {
	// 	return nil, NewError(err, "adding svc ip to wg")
	// }

	if !isFreeIP {
		m.SvcIPCounter++
	}

	return &RegisterServiceResponse{
		ServiceBindingIP:  svcIP,
		ServiceBindingUID: string(svcBinding.GetUID()),
	}, nil
}

func (m *Manager) DeregisterService(ctx context.Context, svcBindingIP string, svcBindingUID string) error {
	var svcBinding networkingv1.ServiceBinding
	if err := m.kcli.Get(ctx, fn.NN("", sanitizeSvcIP(svcBindingIP)), &svcBinding); err != nil {
		return NewError(err, "k8s get svc binding")
	}

	if string(svcBinding.UID) != svcBindingUID {
		m.logger.Info("service binding uid mismatch", "expected", svcBindingUID, "got", svcBinding.UID)
		return nil
	}

	m.FreeSvcIPs = append(m.FreeSvcIPs, svcBindingIP)

	var counterPatch jp.Document
	counterPatch.Add("add", "/data/free_svc_ips", strings.Join(m.FreeSvcIPs, ","))

	b, err := counterPatch.Json()
	if err != nil {
		return NewError(err, "marshalling patches into json array")
	}

	if _, err := m.kclientset.CoreV1().ConfigMaps(m.Env.IPManagerConfigNamespace).Patch(ctx, m.Env.IPManagerConfigName, types.JSONPatchType, b, metav1.PatchOptions{}); err != nil {
		return NewError(err, "patching configmap")
	}

	if err := m.kcli.Delete(ctx, &svcBinding); err != nil {
		return NewError(err, "k8s delete svc binding")
	}

	// if err := m.WgRemoveAddr(svcBindingIP); err != nil {
	// 	return NewError(err, "removing ip addr from wg")
	// }

	if err := m.DeregisterAndSyncNginxStreams(ctx, svcBindingIP); err != nil {
		return NewError(err, "deregistering nginx stream")
	}

	return nil
}
