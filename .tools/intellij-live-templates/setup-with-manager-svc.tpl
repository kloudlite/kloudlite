	builder := ctrl.NewControllerManagedBy(mgr).For(&$Package$.$Type${})

	builder.Owns(fn.NewUnstructured(constants.$TypeMeta$))
	builder.Owns(&corev1.Secret{})

	refWatchList := []client.Object{
		&corev1.Pod{},
	}

	for _, item := range refWatchList {
		builder.Watches(
			&source.Kind{Type: item}, handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					value, ok := obj.GetLabels()[fmt.Sprintf("%s/ref", $Package$.GroupVersion.Group)]
					if !ok {
						return nil
					}
					return []reconcile.Request{
						{NamespacedName: fn.NN(obj.GetNamespace(), value)},
					}
				},
			),
		)
	}

	return builder.Complete(r)

