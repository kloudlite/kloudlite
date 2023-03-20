{{ $gitlabUrl := get . "gitlab-url" |default "https://gitlab.com" }} 
{{ $runnerToken := get . "runner-token" }} 
gitlabUrl: {{$gitlabUrl}}
runnerRegistrationToken:  {{$runnerToken}}
rbac:
  create: true
