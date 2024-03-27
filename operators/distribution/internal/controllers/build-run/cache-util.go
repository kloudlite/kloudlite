package buildrun

import (
	"fmt"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

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
	for i, path := range obj.Spec.Caches {
		name := path.Name
		containerName := fmt.Sprintf("tmp_container_%d", i)

		checkoutCmd += fmt.Sprintf(`
docker pull %s:%s || true
if [ "$(docker images -q %s:%s)" == "" ]; then
    echo "[#] image %s:%s not found, using busybox instead."
    docker create --name %s alpine:latest
else
    docker create --name %s %s:%s
fi
docker cp %s:/data %q || mkdir -p %q
		  `,
			repoName, name,
			repoName, name,
			repoName, name,
			containerName,
			containerName, repoName, name,
			containerName, path, path,
		)

		postCheckoutCmd += fmt.Sprintf(`
mkdir -p %q
docker cp %q %s:/data || true
docker commit %s %s:%s
docker push %s:%s
		`,
			path,
			path, containerName,
			containerName, repoName, name,
			repoName, name,
		)
	}

	return checkoutCmd, postCheckoutCmd, nil
}
