import * as React from 'react'
import { Illustration } from '@kloudlite/ui'

export const Resting = () => (
  <div
    className="w-96 p-4"
    style={{ background: 'linear-gradient(110deg,#333 0.6%,#222)' }}
  >
    <Illustration mouseEnter={false} />
  </div>
)

export const AsCardHeader = () => (
  <div
    className="w-96 p-4"
    style={{ background: 'linear-gradient(110deg,#333 0.6%,#222)' }}
  >
    <Illustration mouseEnter={false} />
    <div className="px-2 pb-6">
      <h2 className="text-2xl font-bold" style={{ color: '#eaeaea' }}>
        Ship from anywhere
      </h2>
      <p className="text-base" style={{ color: '#ffffff' }}>
        Kloudlite connects your laptop to a live Kubernetes environment.
      </p>
    </div>
  </div>
)
