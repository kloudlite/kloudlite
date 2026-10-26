import * as React from 'react'
import {
  GlowingStarsBackgroundCard,
  GlowingStarsDescription,
  GlowingStarsTitle,
} from '@kloudlite/ui'

export const InCard = () => (
  <div className="w-96" style={{ height: '20rem' }}>
    <GlowingStarsBackgroundCard>
      <GlowingStarsTitle>Ship from anywhere</GlowingStarsTitle>
      <GlowingStarsDescription>
        Kloudlite connects your laptop to a live Kubernetes environment.
      </GlowingStarsDescription>
    </GlowingStarsBackgroundCard>
  </div>
)

export const TitleOnly = () => (
  <div className="w-96" style={{ height: '20rem' }}>
    <GlowingStarsBackgroundCard>
      <GlowingStarsTitle>Environments in seconds</GlowingStarsTitle>
    </GlowingStarsBackgroundCard>
  </div>
)
