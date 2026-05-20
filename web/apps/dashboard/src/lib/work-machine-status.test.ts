import { describe, expect, test } from 'vitest'
import { getCondition, isWorkMachineReady, workMachineStatusMessage, workMachineStatusReason } from './work-machine-status'

describe('work-machine-status helpers', () => {
  test('returns ready when the Ready condition status is True', () => {
    const workMachine = {
      status: {
        conditions: [
          { type: 'Provisioned', status: 'False' },
          { type: 'Ready', status: 'True', message: 'workspace agent is ready' },
        ],
      },
    }

    expect(isWorkMachineReady(workMachine)).toBe(true)
    expect(workMachineStatusMessage(workMachine)).toBe('workspace agent is ready')
  })

  test('reads readiness from the Ready condition', () => {
    const workMachine = {
      status: {
        conditions: [
          { type: 'Provisioned', status: 'True' },
          { type: 'Ready', status: 'False', message: 'workspace agent is not ready' },
        ],
      },
    }

    expect(getCondition(workMachine, 'Ready')).toEqual({
      type: 'Ready',
      status: 'False',
      message: 'workspace agent is not ready',
    })
    expect(isWorkMachineReady(workMachine)).toBe(false)
    expect(workMachineStatusMessage(workMachine)).toBe('workspace agent is not ready')
  })

  test('treats missing Ready condition as not ready', () => {
    const workMachine = {
      status: {
        message: 'legacy ready message',
      },
    }

    expect(isWorkMachineReady(workMachine)).toBe(false)
    expect(workMachineStatusMessage(workMachine)).toBe('legacy ready message')
  })

  test('falls back to unavailable when no Ready condition or status message exists', () => {
    expect(workMachineStatusMessage({ status: {} })).toBe('Status unavailable')
  })

  test('exposes the Ready condition reason', () => {
    const workMachine = {
      status: {
        conditions: [
          { type: 'Ready', status: 'False', reason: 'MachineWorkloadsNotReady', message: 'workloads pending' },
        ],
      },
    }

    expect(workMachineStatusReason(workMachine)).toBe('MachineWorkloadsNotReady')
  })
})
