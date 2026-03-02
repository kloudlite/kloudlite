import { getRazorpay } from './razorpay'
import { getPlans, updatePlanRazorpayId } from './console/storage'

/**
 * Ensure all subscription plans have corresponding Razorpay plans.
 * Creates missing plans in Razorpay and updates our DB with the Razorpay plan ID.
 */
export async function syncPlansToRazorpay(): Promise<void> {
  const razorpay = getRazorpay()
  const plans = await getPlans()

  for (const plan of plans) {
    if (plan.razorpayPlanId) continue

    const razorpayPlan = await razorpay.plans.create({
      period: 'monthly',
      interval: 1,
      item: {
        name: `Kloudlite ${plan.name} — ${plan.cpu} vCPU, ${plan.ram} (per user)`,
        amount: plan.amountPerUser,
        currency: plan.currency,
        description: plan.description || `Kloudlite ${plan.name} compute`,
      },
    })

    await updatePlanRazorpayId(plan.id, razorpayPlan.id)
    console.log(`Synced plan ${plan.name} → Razorpay plan ${razorpayPlan.id}`)
  }
}
