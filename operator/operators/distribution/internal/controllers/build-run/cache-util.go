package buildrun

import (
	"fmt"
	"strings"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

const postScript = `
# check if path exists
if [ -d {{path}} ]; then
    docker cp {{path}} {{containerName}}:/data
    docker commit {{containerName}} {{repoName}}:{{tag}}
    docker push {{repoName}}:{{tag}}
else
    echo "[#] cache path {{path}} not found, skipping..."
fi
`

const preScript = `
docker pull {{repoName}}:{{tag}} || true
if [ "$(docker images -q {{repoName}}:{{tag}})" == "" ]; then
    echo "[#] image {{repoName}}:{{tag}} not found, using new image instead."
    docker create --name {{containerName}} alpine:latest
else
    docker create --name {{containerName}} {{repoName}}:{{tag}}
fi
docker cp {{containerName}}:/data {{path}} || echo "[#] failed to copy cache path {{path}}"
`

func parseScript(script string, data map[string]string) string {
	for k, v := range data {
		script = strings.ReplaceAll(script, fmt.Sprintf("{{%s}}", k), v)
	}
	return script
}

func (r *Reconciler) getCacheCmds(req *rApi.Request[*dbv1.BuildRun]) (string, string, error) {
	obj := req.Object

	if obj.Spec.Caches == nil || len(obj.Spec.Caches) == 0 {
		return "", "", nil
	}

	err, _, _, rh, _ := r.getCreds(req)
	if err != nil {
		return "", "", err
	}

	repoName := fmt.Sprintf("%s/%s/%s", rh, obj.Spec.AccountName, "kloudlite-cache")

	checkoutCmd := "echo '[#] checking out cache paths...'"
	postCheckoutCmd := "echo '[#] pushing cache paths...'"
	for i, cache := range obj.Spec.Caches {
		name := cache.Name
		path := cache.Path

		containerName := fmt.Sprintf("tmp_container_%d", i)

		data := map[string]string{
			"repoName":      repoName,
			"tag":           name,
			"containerName": containerName,
			"path":          fmt.Sprintf("%q", path),
		}

		checkoutCmd += parseScript(preScript, data)
		postCheckoutCmd += parseScript(postScript, data)

	}

	return checkoutCmd, postCheckoutCmd, nil
}
