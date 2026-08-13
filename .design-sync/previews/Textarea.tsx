import * as React from 'react'
import { Label, Textarea } from '@kloudlite/ui'

const MANIFEST = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: kloudlite-dev
spec:
  replicas: 3`

export const Basic = () => (
  <div className="w-96">
    <Textarea defaultValue={MANIFEST} rows={7} />
  </div>
)

export const WithLabel = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="notes">Deployment notes</Label>
    <Textarea
      id="notes"
      rows={5}
      defaultValue={
        'Rolling out the new ingress controller to ap-south-1.\nDrain node pool "default" before applying.\nRollback plan: revert to revision 14.'
      }
    />
  </div>
)

export const Placeholder = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="reason">Reason for access</Label>
    <Textarea id="reason" rows={4} placeholder="Describe why you need production access…" />
  </div>
)

export const Disabled = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="locked">Cluster spec (read only)</Label>
    <Textarea id="locked" rows={7} disabled defaultValue={MANIFEST} />
  </div>
)

export const Invalid = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="bad">Environment variables</Label>
    <Textarea
      id="bad"
      rows={4}
      aria-invalid
      className="border-destructive"
      defaultValue={'DATABASE_URL=postgres://db:5432\nAPI KEY=missing-equals'}
    />
    <p className="text-xs text-destructive">Line 2 is not a valid KEY=VALUE pair.</p>
  </div>
)
