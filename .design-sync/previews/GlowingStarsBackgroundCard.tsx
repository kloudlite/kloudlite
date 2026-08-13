import * as React from 'react'
import {
  GlowingStarsBackgroundCard,
  GlowingStarsDescription,
  GlowingStarsTitle,
} from '@kloudlite/ui'

export const Default = () => (
  <div className="w-96" style={{ height: '20rem' }}>
    <GlowingStarsBackgroundCard>
      <GlowingStarsTitle>Ship from anywhere</GlowingStarsTitle>
      <GlowingStarsDescription>
        Kloudlite connects your laptop to a live Kubernetes environment.
      </GlowingStarsDescription>
    </GlowingStarsBackgroundCard>
  </div>
)

export const FeatureTile = () => (
  <div className="w-96" style={{ height: '20rem' }}>
    <GlowingStarsBackgroundCard>
      <GlowingStarsTitle>Zero-config networking</GlowingStarsTitle>
      <GlowingStarsDescription>
        Reach every service in your cluster by its in-cluster hostname, no port
        forwards.
      </GlowingStarsDescription>
    </GlowingStarsBackgroundCard>
  </div>
)
