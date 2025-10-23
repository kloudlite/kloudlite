'use server'

import { workMachineService } from '@/lib/services/work-machine.service'

export async function getMyWorkMachine() {
  try {
    const data = await workMachineService.getMyWorkMachine()
    return { success: true, data }
  } catch (err) {
    console.error('Get my work machine error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

export async function listAllWorkMachines() {
  try {
    const data = await workMachineService.listAllWorkMachines()
    return { success: true, data }
  } catch (err) {
    console.error('List work machines error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

export async function startMyWorkMachine() {
  try {
    const data = await workMachineService.startMyWorkMachine()
    return { success: true, data }
  } catch (err) {
    console.error('Start work machine error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

export async function stopMyWorkMachine() {
  try {
    const data = await workMachineService.stopMyWorkMachine()
    return { success: true, data }
  } catch (err) {
    console.error('Stop work machine error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

export async function updateMyWorkMachine(updateData: {
  machineType?: string
  sshPublicKeys?: string[]
}) {
  try {
    const data = await workMachineService.updateMyWorkMachine(updateData)
    return { success: true, data }
  } catch (err) {
    console.error('Update work machine error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}
