package conditions

const (
	HelmResourceExists     Type = "HelmResourceExists"
	DeploymentExists       Type = "DeploymentExists"
	StsExists              Type = "StsExists"
	GeneratedVars          Type = "GeneratedVars"
	ReconcilerOutputExists Type = "ReconcilerOutputExists"

	ManagedSvcExists       Type = "ManagedSvcExists"
	ManagedSvcOutputExists Type = "ManagedSvcOutputExists"

	// ---

	HelmResourceReady Type = "HelmResourceReady"
	DeploymentReady   Type = "DeploymentReady"
	StsReady          Type = "StsReady"
	OutputReady       Type = "OutputReady"
	ManagedSvcReady   Type = "ManagedSvcReady"
)

const (
	Found             Reason = "Found"
	NotFound          Reason = "NotFound"
	NotReconciledYet  Reason = "NotReconciledYet"
	ErrWhileReconcile Reason = "ErrWhileReconcilation"
	Empty             Reason = ""
)
