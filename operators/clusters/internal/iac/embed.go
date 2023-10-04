package iac

import "embed"

//go:embed templates/*
var TemplatesDir embed.FS

func ClusterPlanAndApplyTemplate() ([]byte, error) {
	return TemplatesDir.ReadFile("templates/cluster-plan-and-apply-job.yml.tpl")
}

func ClusterDestroyJobTemplate() ([]byte, error) {
	return TemplatesDir.ReadFile("templates/cluster-destroy-job.yml.tpl")
}

func ClusterJobRBACTemplate() ([]byte, error) {
	return TemplatesDir.ReadFile("templates/cluster-job-rbac.yml.tpl")
}
