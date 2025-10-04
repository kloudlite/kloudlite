'use server'

import { workMachineService } from '@/lib/services/work-machine.service'

export async function getMyWorkMachine() {
  try {
    const data = await workMachineService.getMyWorkMachine()
    return { success: true, data }
  } catch (error: any) {
    console.error('Get my work machine error:', error)
    return {
      success: false,
      error: error.message || 'Failed to get work machine'
    }
  }
}

export async function listAllWorkMachines() {
  try {
    const data = await workMachineService.listAllWorkMachines()
    return { success: true, data }
  } catch (error: any) {
    console.error('List work machines error:', error)
    return {
      success: false,
      error: error.message || 'Failed to list work machines'
    }
  }
}

export async function startMyWorkMachine() {
  try {
    const data = await workMachineService.startMyWorkMachine()
    return { success: true, data }
  } catch (error: any) {
    console.error('Start work machine error:', error)
    return {
      success: false,
      error: error.message || 'Failed to start work machine'
    }
  }
}

export async function stopMyWorkMachine() {
  try {
    const data = await workMachineService.stopMyWorkMachine()
    return { success: true, data }
  } catch (error: any) {
    console.error('Stop work machine error:', error)
    return {
      success: false,
      error: error.message || 'Failed to stop work machine'
    }
  }
}

export async function updateMyWorkMachine(updateData: { machineType?: string; sshPublicKeys?: string[] }) {
  try {
    const data = await workMachineService.updateMyWorkMachine(updateData)
    return { success: true, data }
  } catch (error: any) {
    console.error('Update work machine error:', error)
    return {
      success: false,
      error: error.message || 'Failed to update work machine'
    }
  }
}
