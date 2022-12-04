package constants

const (
	NsCore        = "kl-core"
	NsRedpanda    = "kl-init-redpanda"
	NsIngress     = "kl-init-ingress"
	NsCertManager = "kl-init-cert-manager"
	NsMonitoring  = "kl-init-monitoring"
	NsHarbor      = "kl-init-harbor"
	NsOperators   = "kl-init-operators"
)

const (
	DefaultPullSecret = "kloudlite-docker-registry"
	DefaultSvcAccount = "kloudlite-svc-account"
	ClusterSvcAccount = "kloudlite-cluster-svc-account"
	//DefaultCertIssuerName = "kl-cert-issuer"
	//WildcardTLSSecretName = "kl-cert-issuer-tls"
)
