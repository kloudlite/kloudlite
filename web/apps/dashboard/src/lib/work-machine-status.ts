type WorkMachineCondition = {
  type?: string
  status?: string
  reason?: string
  message?: string
}

type WorkMachineWithConditions = {
  status?: unknown
}

export function getCondition(workMachine: WorkMachineWithConditions | null | undefined, type: string) {
  const status = workMachine?.status
  if (typeof status !== 'object' || status === null) return undefined

  const conditions = (status as { conditions?: unknown }).conditions
  if (!Array.isArray(conditions)) return undefined

  return conditions.find((condition): condition is WorkMachineCondition => {
    return typeof condition === 'object' && condition !== null && (condition as WorkMachineCondition).type === type
  })
}

export function isWorkMachineReady(workMachine: WorkMachineWithConditions | null | undefined) {
  return getCondition(workMachine, 'Ready')?.status === 'True'
}

export function workMachineStatusMessage(workMachine: WorkMachineWithConditions | null | undefined) {
  const status = workMachine?.status
  const statusMessage = typeof status === 'object' && status !== null ? (status as { message?: unknown }).message : undefined
  return getCondition(workMachine, 'Ready')?.message ?? (typeof statusMessage === 'string' ? statusMessage : undefined) ?? 'Status unavailable'
}

export function workMachineStatusReason(workMachine: WorkMachineWithConditions | null | undefined) {
  return getCondition(workMachine, 'Ready')?.reason
}
