import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { computeProration } from './proration'
import type { Plan } from '@/lib/console/storage'

const NOW = '2025-01-15T00:00:00Z'
const NOW_MS = new Date(NOW).getTime()

function makePlan(overrides: Partial<Plan> = {}): Plan {
  return {
    id: 'plan-1',
    razorpayPlanId: null,
    tier: 1,
    name: 'Standard',
    amountPerUser: 10000,
    baseFee: 50000,
    currency: 'INR',
    monthlyHours: 720,
    overageRate: 0,
    cpu: 2,
    ram: '4GB',
    storage: '20GB',
    autoSuspend: '30m',
    description: null,
    annualDiscountPct: 20,
    createdAt: '2025-01-01T00:00:00Z',
    ...overrides,
  }
}

type SubInput = {
  planId?: string
  quantity?: number
  billingPeriod?: 'monthly' | 'annual'
  currentEnd?: string | null
}

function makeSub(overrides: SubInput = {}) {
  return {
    planId: 'plan-1',
    quantity: 3,
    billingPeriod: 'monthly' as const,
    currentEnd: '2025-01-30T00:00:00Z',
    ...overrides,
  }
}

describe('computeProration', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(NOW))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns zeros when currentEnd is null', () => {
    const result = computeProration(
      [makeSub({ currentEnd: null })],
      [{ planId: 'plan-1', quantity: 5 }],
      [makePlan()],
    )

    expect(result).toEqual({
      proratedAmount: 0,
      remainingDays: 0,
      oldMonthlyTotal: 0,
      newMonthlyTotal: 0,
      newCurrentEnd: null,
    })
  })

  it('returns zeros when activeSubs is empty (no currentEnd)', () => {
    const result = computeProration(
      [],
      [{ planId: 'plan-1', quantity: 5 }],
      [makePlan()],
    )

    expect(result).toEqual({
      proratedAmount: 0,
      remainingDays: 0,
      oldMonthlyTotal: 0,
      newMonthlyTotal: 0,
      newCurrentEnd: null,
    })
  })

  describe('monthly same-period changes', () => {
    it('prorates a quantity upgrade (3→5 users, 15 days left)', () => {
      // oldUserTotal = 10000*3 = 30000, newUserTotal = 10000*5 = 50000
      // oldPeriodTotal = 50000+30000 = 80000, newPeriodTotal = 50000+50000 = 100000
      // dailyDiff = (100000 - 80000) / 30 = 666.666...
      // prorated = Math.round(666.666... * 15) = 10000
      const result = computeProration(
        [makeSub({ quantity: 3 })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
      )

      expect(result).toEqual({
        proratedAmount: 10000,
        remainingDays: 15,
        oldMonthlyTotal: 30000,
        newMonthlyTotal: 50000,
        newCurrentEnd: null,
      })
    })

    it('returns negative prorated amount for a downgrade (5→2 users)', () => {
      // oldUserTotal = 50000, newUserTotal = 20000
      // oldPeriodTotal = 100000, newPeriodTotal = 70000
      // dailyDiff = (70000 - 100000) / 30 = -1000
      // prorated = Math.round(-1000 * 15) = -15000
      const result = computeProration(
        [makeSub({ quantity: 5 })],
        [{ planId: 'plan-1', quantity: 2 }],
        [makePlan()],
      )

      expect(result).toEqual({
        proratedAmount: -15000,
        remainingDays: 15,
        oldMonthlyTotal: 50000,
        newMonthlyTotal: 20000,
        newCurrentEnd: null,
      })
    })

    it('returns zero prorated amount when quantity is unchanged', () => {
      const result = computeProration(
        [makeSub({ quantity: 3 })],
        [{ planId: 'plan-1', quantity: 3 }],
        [makePlan()],
      )

      expect(result).toEqual({
        proratedAmount: 0,
        remainingDays: 15,
        oldMonthlyTotal: 30000,
        newMonthlyTotal: 30000,
        newCurrentEnd: null,
      })
    })
  })

  describe('annual same-period changes', () => {
    it('prorates an annual quantity upgrade (3→5 users, 180 days left)', () => {
      // currentEnd = 2025-07-14 → 180 days from Jan 15
      // oldUserTotal = 30000, newUserTotal = 50000
      // oldPeriodTotal = (50000+30000)*12*0.8 = 768000
      // newPeriodTotal = (50000+50000)*12*0.8 = 960000
      // dailyDiff = (960000-768000)/365 = 192000/365 = 526.02739726...
      // prorated = Math.round(526.02739726... * 180) = Math.round(94684.93...) = 94685
      const result = computeProration(
        [makeSub({ quantity: 3, billingPeriod: 'annual', currentEnd: '2025-07-14T00:00:00Z' })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
      )

      expect(result).toEqual({
        proratedAmount: 94685,
        remainingDays: 180,
        oldMonthlyTotal: 30000,
        newMonthlyTotal: 50000,
        newCurrentEnd: null,
      })
    })
  })

  describe('period switches', () => {
    it('prorates monthly → annual switch (same quantity)', () => {
      // oldPeriodTotal = 50000+30000 = 80000 (monthly)
      // newPeriodTotal = (50000+30000)*12*0.8 = 768000 (annual)
      // oldDailyRate = 80000/30 = 2666.666...
      // remainingCredit = 2666.666... * 15 = 40000
      // prorated = Math.round(768000 - 40000) = 728000
      // newCurrentEnd = 2026-01-15T00:00:00.000Z (now + 365 days)
      const result = computeProration(
        [makeSub({ quantity: 3 })],
        [{ planId: 'plan-1', quantity: 3 }],
        [makePlan()],
        'annual',
      )

      expect(result).toEqual({
        proratedAmount: 728000,
        remainingDays: 15,
        oldMonthlyTotal: 30000,
        newMonthlyTotal: 30000,
        newCurrentEnd: new Date(NOW_MS + 365 * 24 * 60 * 60 * 1000).toISOString(),
      })
    })

    it('prorates monthly → annual with quantity upgrade simultaneously', () => {
      // oldPeriodTotal = 80000 (monthly, 3 users)
      // newPeriodTotal = (50000+50000)*12*0.8 = 960000 (annual, 5 users)
      // oldDailyRate = 80000/30 = 2666.666...
      // remainingCredit = 2666.666... * 15 = 40000
      // prorated = Math.round(960000 - 40000) = 920000
      const result = computeProration(
        [makeSub({ quantity: 3 })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
        'annual',
      )

      expect(result).toEqual({
        proratedAmount: 920000,
        remainingDays: 15,
        oldMonthlyTotal: 30000,
        newMonthlyTotal: 50000,
        newCurrentEnd: new Date(NOW_MS + 365 * 24 * 60 * 60 * 1000).toISOString(),
      })
    })

    it('prorates annual → monthly switch', () => {
      // activeSub: annual, 180 days left, 3 users
      // oldPeriodTotal = (50000+30000)*12*0.8 = 768000 (annual)
      // newPeriodTotal = 50000+30000 = 80000 (monthly)
      // oldDailyRate = 768000/365 = 2104.10958904...
      // remainingCredit = 2104.10958904... * 180 = 378739.726027...
      // prorated = Math.round(80000 - 378739.726027...) = Math.round(-298739.726...) = -298740
      // newCurrentEnd = now + 365 days (code always uses 365 for period changes)
      const result = computeProration(
        [makeSub({ quantity: 3, billingPeriod: 'annual', currentEnd: '2025-07-14T00:00:00Z' })],
        [{ planId: 'plan-1', quantity: 3 }],
        [makePlan()],
        'monthly',
      )

      expect(result.proratedAmount).toBe(-298740)
      expect(result.remainingDays).toBe(180)
      expect(result.newCurrentEnd).toBe(new Date(NOW_MS + 365 * 24 * 60 * 60 * 1000).toISOString())
    })
  })

  describe('multiple plan tiers', () => {
    it('prorates across two tiers with different amountPerUser', () => {
      const plans = [
        makePlan({ id: 'plan-1', amountPerUser: 10000 }),
        makePlan({ id: 'plan-2', amountPerUser: 20000 }),
      ]
      // oldUserTotal = 10000*2 + 20000*1 = 40000
      // newUserTotal = 10000*3 + 20000*2 = 70000
      // baseFee = plans[0].baseFee = 50000
      // oldPeriodTotal = 50000+40000 = 90000
      // newPeriodTotal = 50000+70000 = 120000
      // dailyDiff = (120000-90000)/30 = 1000
      // prorated = Math.round(1000 * 15) = 15000
      const result = computeProration(
        [
          makeSub({ planId: 'plan-1', quantity: 2 }),
          makeSub({ planId: 'plan-2', quantity: 1 }),
        ],
        [
          { planId: 'plan-1', quantity: 3 },
          { planId: 'plan-2', quantity: 2 },
        ],
        plans,
      )

      expect(result).toEqual({
        proratedAmount: 15000,
        remainingDays: 15,
        oldMonthlyTotal: 40000,
        newMonthlyTotal: 70000,
        newCurrentEnd: null,
      })
    })
  })

  describe('edge cases', () => {
    it('returns 0 prorated amount when currentEnd is in the past', () => {
      // currentEnd = 2025-01-10 → 5 days ago → remainingDays = 0
      const result = computeProration(
        [makeSub({ quantity: 3, currentEnd: '2025-01-10T00:00:00Z' })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
      )

      expect(result.proratedAmount).toBe(0)
      expect(result.remainingDays).toBe(0)
      expect(result.oldMonthlyTotal).toBe(30000)
      expect(result.newMonthlyTotal).toBe(50000)
    })

    it('treats unknown planId as 0 amountPerUser', () => {
      // oldUserTotal = 0 (unknown-plan not in planMap)
      // newUserTotal = 10000*5 = 50000
      // oldPeriodTotal = 50000+0 = 50000, newPeriodTotal = 50000+50000 = 100000
      // dailyDiff = (100000-50000)/30 = 1666.666...
      // prorated = Math.round(1666.666... * 15) = 25000
      const result = computeProration(
        [makeSub({ planId: 'unknown-plan', quantity: 3 })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
      )

      expect(result.proratedAmount).toBe(25000)
      expect(result.oldMonthlyTotal).toBe(0)
      expect(result.newMonthlyTotal).toBe(50000)
    })

    it('uses baseFee in period switch calculations', () => {
      // baseFee = 0: monthly → annual, 3 users, 15 days left
      // oldPeriodTotal = 0+30000 = 30000
      // newPeriodTotal = (0+30000)*12*0.8 = 288000
      // oldDailyRate = 30000/30 = 1000
      // remainingCredit = 1000*15 = 15000
      // prorated = Math.round(288000-15000) = 273000
      const result = computeProration(
        [makeSub({ quantity: 3 })],
        [{ planId: 'plan-1', quantity: 3 }],
        [makePlan({ baseFee: 0 })],
        'annual',
      )

      expect(result.proratedAmount).toBe(273000)
    })

    it('handles annualDiscountPct = 0 (no discount)', () => {
      // Annual upgrade 3→5, 180 days left, no discount
      // oldPeriodTotal = (50000+30000)*12*1.0 = 960000
      // newPeriodTotal = (50000+50000)*12*1.0 = 1200000
      // dailyDiff = (1200000-960000)/365 = 240000/365 = 657.53424657...
      // prorated = Math.round(657.53424657... * 180) = Math.round(118356.16...) = 118356
      const result = computeProration(
        [makeSub({ quantity: 3, billingPeriod: 'annual', currentEnd: '2025-07-14T00:00:00Z' })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan({ annualDiscountPct: 0 })],
      )

      expect(result.proratedAmount).toBe(118356)
      expect(result.remainingDays).toBe(180)
    })

    it('handles currentEnd exactly at now (0 remaining days)', () => {
      const result = computeProration(
        [makeSub({ quantity: 3, currentEnd: NOW })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
      )

      expect(result.proratedAmount).toBe(0)
      expect(result.remainingDays).toBe(0)
    })

    it('handles 1 remaining day', () => {
      // 1 day left: currentEnd = 2025-01-16T00:00:00Z
      // dailyDiff = (100000-80000)/30 = 666.666...
      // prorated = Math.round(666.666... * 1) = 667
      const result = computeProration(
        [makeSub({ quantity: 3, currentEnd: '2025-01-16T00:00:00Z' })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
      )

      expect(result.proratedAmount).toBe(667)
      expect(result.remainingDays).toBe(1)
    })

    it('ceils partial remaining days (currentEnd at noon)', () => {
      // now = Jan 15 00:00, currentEnd = Jan 30 12:00 → 15.5 days → ceil = 16
      // dailyDiff = (100000-80000)/30 = 666.666...
      // prorated = Math.round(666.666... * 16) = Math.round(10666.666...) = 10667
      const result = computeProration(
        [makeSub({ quantity: 3, currentEnd: '2025-01-30T12:00:00Z' })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
      )

      expect(result.remainingDays).toBe(16)
      expect(result.proratedAmount).toBe(10667)
    })

    it('uses first sub currentEnd when multiple subs exist', () => {
      // First sub has currentEnd = Jan 30 (15 days), second has Jan 25 — only first is used
      const result = computeProration(
        [
          makeSub({ planId: 'plan-1', quantity: 2, currentEnd: '2025-01-30T00:00:00Z' }),
          makeSub({ planId: 'plan-1', quantity: 1, currentEnd: '2025-01-25T00:00:00Z' }),
        ],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
      )

      expect(result.remainingDays).toBe(15)
    })

    it('handles adding a new plan tier not in activeSubs', () => {
      const plans = [
        makePlan({ id: 'plan-1', amountPerUser: 10000 }),
        makePlan({ id: 'plan-2', amountPerUser: 20000 }),
      ]
      // oldUserTotal = 10000*3 = 30000 (only plan-1)
      // newUserTotal = 10000*3 + 20000*2 = 70000 (plan-1 + new plan-2)
      // oldPeriodTotal = 50000+30000 = 80000, newPeriodTotal = 50000+70000 = 120000
      // dailyDiff = (120000-80000)/30 = 1333.333...
      // prorated = Math.round(1333.333... * 15) = 20000
      const result = computeProration(
        [makeSub({ planId: 'plan-1', quantity: 3 })],
        [
          { planId: 'plan-1', quantity: 3 },
          { planId: 'plan-2', quantity: 2 },
        ],
        plans,
      )

      expect(result.proratedAmount).toBe(20000)
      expect(result.oldMonthlyTotal).toBe(30000)
      expect(result.newMonthlyTotal).toBe(70000)
    })

    it('oldMonthlyTotal and newMonthlyTotal exclude baseFee', () => {
      const result = computeProration(
        [makeSub({ quantity: 3 })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan({ baseFee: 100000 })],
      )

      expect(result.oldMonthlyTotal).toBe(30000)
      expect(result.newMonthlyTotal).toBe(50000)
    })

    it('newCurrentEnd is null when newBillingPeriod matches existing period', () => {
      const result = computeProration(
        [makeSub({ quantity: 3, billingPeriod: 'monthly' })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
        'monthly',
      )

      expect(result.newCurrentEnd).toBeNull()
    })

    it('newCurrentEnd is null when newBillingPeriod is omitted', () => {
      const result = computeProration(
        [makeSub({ quantity: 3 })],
        [{ planId: 'plan-1', quantity: 5 }],
        [makePlan()],
      )

      expect(result.newCurrentEnd).toBeNull()
    })
  })
})
