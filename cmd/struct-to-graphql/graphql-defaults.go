package main

import "bytes"

func Directives() ([]byte, error) {
	directives := `
extend schema @link(url: "https://specs.apollo.dev/federation/v2.0", import: ["@key", "@shareable"])

directive @goField(
	forceResolver: Boolean
	name: String
) on INPUT_FIELD_DEFINITION | FIELD_DEFINITION
`
	return []byte(directives), nil
}

func ScalarTypes() ([]byte, error) {
	scalars := `
scalar Any
scalar Json
scalar Map
scalar Date
`

	metadata := `
type Metadata @shareable {
	name: String!
	namespace: String
	labels: Json
	annotations: Json
	creationTimestamp: Date!
	deletionTimestamp: Date
	generation: Int!
}

input MetadataIn {
	name: String!
	namespace: String
	labels: Json
	annotations: Json
}
`

	status := `
type Status @shareable {
	isReady: Boolean!
	checks: Map
	displayVars: Json
}

type Check @shareable {
	status: Boolean
	message: String
	generation: Int
}
`

	overrides := `
type Patch @shareable {
	op: String!
	path: String!
	value: Any
}

type Overrides @shareable{
	applied: Boolean
	patches: [Patch!]
}

input PatchIn {
	op: String!
	path: String!
	value: Any
}

input OverridesIn{
	patches: [PatchIn!]
}
`

	syncStatus := `
enum SyncAction {
	APPLY
	DELETE
}

enum SyncState {
	IDLE
	IN_PROGRESS
	READY
	NOT_READY
}

type SyncStatus @shareable{
	syncScheduledAt: Date!
	lastSyncedAt: Date
	action: SyncAction!
	generation: Int!
	state: SyncState!
	error: String
}
`

	b := bytes.NewBuffer(nil)
	b.WriteString(scalars)
	b.WriteString(metadata)
	b.WriteString(status)
	b.WriteString(overrides)
	b.WriteString(syncStatus)

	return b.Bytes(), nil
}
