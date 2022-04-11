package v1

func (p *Project) IsNewGeneration() bool {
	return p.Status.Generation != p.Generation
}
