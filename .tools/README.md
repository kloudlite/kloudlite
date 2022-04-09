
### CRDs

- [x] Project
- [x] App
- [x] Router
- [x] ManagedService
- [x] ManagedResource

### learnings


+ operator-sdk Reconcile()
 
```go 
// With the error:
return ctrl.Result{}, err

// Without an error:
return ctrl.Result{Requeue: true}, nil

// Reconcile again after X time:
return ctrl.Result{RequeueAfter: nextRun.Sub(r.Now())}, nil

// to stop reconciling
return ctrl.Result{}, nil
```

