package kube_client

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	client.Client
}

func (c *Client) FailWithErr(ctx context.Context, obj client.Object, err error) (ctrl.Result, error) {
	//fn.Conditions2.MarkNotReady(&req.database.Status.OpsConditions, err, "ReconFailedWithErr")
	if err := c.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, err
}
