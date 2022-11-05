package controllers

import (
	"fmt"
)

func GetClusterIssuerName(region string) string {
	return fmt.Sprintf("kl-cert-issuer-%s", region)
}

func GetCertIssuerSecret(region string) string {
	return GetClusterIssuerName(region) + "-tls"
}

func GetIngressClassName(region string) string {
	return fmt.Sprintf("ingress-nginx-%s", region)
}
