package buildrun

import (
	"crypto/md5"
	"fmt"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

func (r *Reconciler) getCacheCmds(req *rApi.Request[*dbv1.BuildRun]) (string, string, error) {
	obj := req.Object

	if obj.Spec.CachePaths == nil || len(obj.Spec.CachePaths) == 0 {
		return "", "", nil
	}

	err, _, _, rh, _ := r.getCreds(req)
	if err != nil {
		return "", "", err
	}

	repoName := fmt.Sprintf("%s/%s/%s", rh, obj.Spec.AccountName, "kloudlite-cache")

	checkoutCmd := "echo '[#] checking out cache paths...'"
	postCheckoutCmd := "echo '[#] pushing cache paths...'"
	for i, path := range obj.Spec.CachePaths {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(path)))
		containerName := fmt.Sprintf("tmp_container_%d", i)

		checkoutCmd += fmt.Sprintf(`
docker pull %s:%s || true
if [ "$(docker images -q %s:%s)" == "" ]; then
    echo "[#] image %s:%s not found, using busybox instead."
    docker create --name %s alpine:latest
else
    docker create --name %s %s:%s
fi
docker cp %s:/%s %q || mkdir -p %q
		  `,
			repoName, hash,
			repoName, hash,
			repoName, hash,
			containerName,
			containerName, repoName, hash,
			containerName, hash, path, path,
		)

		postCheckoutCmd += fmt.Sprintf(`
mkdir -p %q
docker cp %q %s:/%s || true
docker commit %s %s:%s
docker push %s:%s
		`,
			path,
			path, containerName, hash,
			containerName, repoName, hash,
			repoName, hash,
		)
	}

	return checkoutCmd, postCheckoutCmd, nil
}
