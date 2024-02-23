#!/usr/bin/env zx

$.verbose = true

async function checkReleaseExists(githubRepo, release) {
  try {
    await $`gh release list -R ${githubRepo} | tail -n +1 | awk '{print $3}' | grep -iE '^${release}$'`;
  } catch(err) {
    return false
  }
  return true
}

async function main() {
  const releaseTag = process.env.RELEASE_TAG
  const githubRepo = process.env.GITHUB_REPOSITORY
  const githubRef = process.env.GITHUB_REF
  const files = (process.env.FILES || "").split("\n").filter(item => item.length > 0)

  const isReleaseFromBranch = githubRef.startsWith("refs/heads/")
  const branchName = githubRef.replace("refs/heads/", "")

  if (await checkReleaseExists(githubRepo, releaseTag)) {
    if (!isReleaseFromBranch) {
      console.log(`RELEASE (${releaseTag}) already for git tag (${githubRef})`)
      process.exit(1)
    }
    console.log("RELEASE exists, need to delete current release, first")
    await $`gh release delete ${releaseTag} -R ${githubRepo} -y --cleanup-tag`
    console.log(`release ${releaseTag} deleted, successfully`)
  }

  let releaseOpts = ["--verify-tag"] // defaults to release from a tag
  if (isReleaseFromBranch) {
    releaseOpts = [
      "--target", branchName,
      "--prerelease", // any release created from a branch is a prerelease
    ]
  } 

  const uploadables = await glob(files)
  console.log(`uploading ${uploadables}`)

  // create a new release
  await $`gh release create ${releaseTag} ${releaseOpts} -R ${githubRepo} --generate-notes ${uploadables}`
}

try {
  await main()
} catch (err) {
  console.error(err)
  process.exit(1)
}
