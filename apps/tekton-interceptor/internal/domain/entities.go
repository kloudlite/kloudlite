package domain

type GitPipeline struct {
	GitRepo          string
	GitUser          string
	GitPassword      string
	GitRef           string
	GitCommitHash    string
	Dockerfile       string
	DockerContextDir string
	DockerBuildArgs  string
	DockerImageName  string
	DockerImageTag   string
}
