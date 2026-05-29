# Untrack Generated CRDs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove generated CRD/manifests artifacts from git and generate them during build tasks instead.

**Architecture:** Keep source-of-truth CRD Go types and generator code tracked. Ignore generated YAML/Go/bundled CRD artifacts, make build tasks run `task api:manifests` before compiling binaries that import embedded CRD assets, and remove tracked generated files with `git rm --cached`.

**Tech Stack:** Taskfile v3, controller-gen, Go generate-style helper, git ignore rules.

---

## Files

- Modify `.gitignore`: ignore generated CRD artifacts.
- Modify `api/Taskfile.yml`: make relevant build tasks depend on `manifests`.
- Remove from git index:
  - `manifests/*.yaml`
  - `api/crds/assets/generated.go`
  - `cli/kli/internal/manifests/crds.yaml`
- Keep tracked:
  - `api/crds/assets/generate.go`
  - `api/crds/installer.go`
  - `types/**`

## Tasks

- [ ] Add ignore patterns for generated CRD outputs.
- [ ] Update `api:build:api-server`, `api:build:workmachine-manager`, and `api:build:kloudlite` to run `task: manifests` before `go build`.
- [ ] Remove generated files from git tracking while keeping local files available.
- [ ] Run `task api:manifests`.
- [ ] Run focused Go verification.
- [ ] Commit and push to PR #505.

## Self-review

- Spec coverage: ignores, untracking, build-time generation, verification, and PR push are covered.
- Placeholder scan: no placeholders.
- Path consistency: generated artifact paths match current Taskfile outputs.
