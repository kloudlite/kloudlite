'use client'

import { useState } from 'react'
import { Button, Badge, Card, CardContent, CardHeader, CardTitle } from '@kloudlite/ui'
import { Check, Loader2 } from 'lucide-react'
import type { Plan, Subscription } from '@/lib/console/storage'

interface PlanCardsProps {
  plans: Plan[]
  subscription: Subscription | null
  installationId: string
  isOwner: boolean
  onSubscribe: (planId: string) => Promise<void>
}

export function PlanCards({ plans, subscription, isOwner, onSubscribe }: PlanCardsProps) {
  const [loadingPlanId, setLoadingPlanId] = useState<string | null>(null)

  const activePlanId = subscription && !['cancelled', 'expired'].includes(subscription.status)
    ? subscription.planId
    : null

  const handleSubscribe = async (planId: string) => {
    setLoadingPlanId(planId)
    try {
      await onSubscribe(planId)
    } finally {
      setLoadingPlanId(null)
    }
  }

  const tierFeatures: Record<number, string[]> = {
    1: ['8 vCPU per workspace', '16 GB RAM', '100 GB storage', '15 min auto-suspend', '160 hrs/month included'],
    2: ['12 vCPU per workspace', '32 GB RAM', '200 GB storage', '30 min auto-suspend', '160 hrs/month included'],
    3: ['16 vCPU per workspace', '64 GB RAM', '500 GB storage', '1 hr auto-suspend', '160 hrs/month included'],
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {plans.map((plan) => {
        const isActive = activePlanId === plan.id
        const features = tierFeatures[plan.tier] || []

        return (
          <Card
            key={plan.id}
            className={`relative ${isActive ? 'border-primary ring-1 ring-primary' : 'border-foreground/10'} ${plan.tier === 2 ? 'md:scale-105 md:z-10' : ''}`}
          >
            {plan.tier === 2 && (
              <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                <Badge className="bg-primary text-primary-foreground">Most Popular</Badge>
              </div>
            )}
            {isActive && (
              <div className="absolute -top-3 right-4">
                <Badge variant="outline" className="border-primary text-primary bg-primary/5">
                  Current Plan
                </Badge>
              </div>
            )}
            <CardHeader className="pb-4">
              <CardTitle className="text-lg">{plan.name}</CardTitle>
              <div className="mt-2">
                <span className="text-3xl font-bold">${plan.amountPerUser / 100}</span>
                <span className="text-muted-foreground text-sm">/user/month</span>
              </div>
              <p className="text-muted-foreground text-xs mt-1">
                + ${plan.baseFee / 100}/mo base fee
              </p>
            </CardHeader>
            <CardContent>
              <ul className="space-y-2 mb-6">
                {features.map((feature) => (
                  <li key={feature} className="flex items-center gap-2 text-sm">
                    <Check className="h-4 w-4 text-primary flex-shrink-0" />
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>
              {isOwner && !isActive && (
                <Button
                  className="w-full"
                  variant={plan.tier === 2 ? 'default' : 'outline'}
                  disabled={!!loadingPlanId}
                  onClick={() => handleSubscribe(plan.id)}
                >
                  {loadingPlanId === plan.id ? (
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  ) : null}
                  {activePlanId ? 'Switch Plan' : 'Subscribe'}
                </Button>
              )}
              {isActive && (
                <Button className="w-full" variant="outline" disabled>
                  Active
                </Button>
              )}
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
