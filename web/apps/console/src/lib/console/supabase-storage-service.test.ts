import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { Mock } from 'vitest'
import {
  getUserByEmail,
  saveUserRegistration,
  markInstallationComplete,
  addOrUpdateIpRecord,
  isSubdomainAvailable,
  reserveSubdomain,
  deleteIpRecords,
  deleteDomainReservation,
  resetUserInstallation,
  type UserRegistration,
  type IPRecord,
} from './supabase-storage-service'

// Type for Supabase query response
interface SupabaseResponse<T> {
  data: T | null
  error: Error | null
  count?: number
}

// Type for callback functions in mock chains
type PromiseCallback<T> = (
  value: SupabaseResponse<T>,
) => SupabaseResponse<T> | PromiseLike<SupabaseResponse<T>>

// Create a mockSupabase object that we'll configure in each test
const mockSupabaseClient = {
  from: vi.fn(),
}

// Mock the Supabase module
vi.mock('./supabase', () => ({
  supabase: mockSupabaseClient,
}))

describe('Supabase Storage Service', () => {
  let mockFrom: Mock
  let mockSelect: Mock
  let mockEq: Mock
  let mockSingle: Mock
  let mockUpsert: Mock
  let mockUpdate: Mock
  let mockInsert: Mock
  let mockDelete: Mock

  beforeEach(async () => {
    vi.clearAllMocks()

    // Set up mock chain with proper method returns
    mockSingle = vi.fn().mockResolvedValue({ data: null, error: null })
    mockEq = vi.fn().mockReturnValue({
      single: mockSingle,
      then: <T>(fn: PromiseCallback<T>) => mockSingle.then(fn), // Make eq() thenable for direct resolution
    })
    mockSelect = vi.fn().mockImplementation((columns?: string, options?: { head?: boolean }) => {
      if (options?.head) {
        return Promise.resolve({ data: null, error: null, count: 0 })
      }
      return { eq: mockEq, single: mockSingle }
    })
    mockUpsert = vi.fn().mockResolvedValue({ error: null })
    mockUpdate = vi.fn().mockReturnValue({
      eq: mockEq,
      then: <T>(fn: PromiseCallback<T>) => Promise.resolve({ error: null }).then(fn), // Make update() thenable
    })
    mockInsert = vi.fn().mockReturnValue({ select: mockSelect })
    mockDelete = vi.fn().mockReturnValue({
      eq: mockEq,
      then: <T>(fn: PromiseCallback<T>) => Promise.resolve({ error: null }).then(fn), // Make delete() thenable
    })
    mockFrom = vi.fn().mockReturnValue({
      select: mockSelect,
      upsert: mockUpsert,
      update: mockUpdate,
      insert: mockInsert,
      delete: mockDelete,
    })

    const { supabase } = await import('./supabase')
    supabase.from = mockFrom
  })

  describe('getUserByEmail', () => {
    it('should return user when found', async () => {
      const mockUser = {
        user_id: 'github-123',
        email: 'test@example.com',
        name: 'Test User',
        providers: ['github'],
        registered_at: '2025-01-01T00:00:00Z',
        installation_key: 'key-123',
        secret_key: 'secret-123',
        has_completed_installation: true,
        subdomain: 'test',
        reserved_at: '2025-01-01T00:00:00Z',
        deployment_ready: true,
        last_health_check: '2025-01-01T00:00:00Z',
      }

      mockSingle.mockResolvedValueOnce({ data: mockUser, error: null })
      mockSelect.mockResolvedValueOnce({ data: [], error: null })

      const result = await getUserByEmail('test@example.com')

      expect(result).toEqual({
        userId: 'github-123',
        email: 'test@example.com',
        name: 'Test User',
        providers: ['github'],
        registeredAt: '2025-01-01T00:00:00Z',
        installationKey: 'key-123',
        secretKey: 'secret-123',
        hasCompletedInstallation: true,
        subdomain: 'test',
        reservedAt: '2025-01-01T00:00:00Z',
        deploymentReady: true,
        lastHealthCheck: '2025-01-01T00:00:00Z',
        ipRecords: [],
      })
    })

    it('should return null when user not found', async () => {
      mockSingle.mockResolvedValueOnce({
        data: null,
        error: { code: 'PGRST116' },
      })

      const result = await getUserByEmail('nonexistent@example.com')
      expect(result).toBeNull()
    })

    it('should return user with IP records', async () => {
      const mockUser = {
        user_id: 'github-123',
        email: 'test@example.com',
        name: 'Test User',
        providers: ['github'],
        registered_at: '2025-01-01T00:00:00Z',
        installation_key: 'key-123',
        has_completed_installation: false,
      }

      const mockIpRecords = [
        {
          type: 'installation',
          ip: '1.2.3.4',
          configured_at: '2025-01-01T00:00:00Z',
          dns_record_ids: ['dns-1'],
        },
        {
          type: 'workmachine',
          ip: '1.2.3.5',
          work_machine_name: 'dev1',
          configured_at: '2025-01-01T00:00:00Z',
          dns_record_ids: ['dns-2', 'dns-3'],
        },
      ]

      mockSingle.mockResolvedValueOnce({ data: mockUser, error: null })
      mockSelect.mockResolvedValueOnce({ data: mockIpRecords, error: null })

      const result = await getUserByEmail('test@example.com')

      expect(result?.ipRecords).toHaveLength(2)
      expect(result?.ipRecords[0]).toEqual({
        type: 'installation',
        ip: '1.2.3.4',
        configuredAt: '2025-01-01T00:00:00Z',
        dnsRecordIds: ['dns-1'],
      })
      expect(result?.ipRecords[1]).toEqual({
        type: 'workmachine',
        ip: '1.2.3.5',
        workMachineName: 'dev1',
        configuredAt: '2025-01-01T00:00:00Z',
        dnsRecordIds: ['dns-2', 'dns-3'],
      })
    })
  })

  describe('saveUserRegistration', () => {
    it('should save user registration successfully', async () => {
      mockUpsert.mockResolvedValueOnce({ error: null })

      const registration: UserRegistration = {
        userId: 'github-123',
        email: 'test@example.com',
        name: 'Test User',
        providers: ['github'],
        registeredAt: '2025-01-01T00:00:00Z',
        installationKey: 'key-123',
        hasCompletedInstallation: false,
      }

      await expect(saveUserRegistration(registration)).resolves.toBeUndefined()

      expect(mockFrom).toHaveBeenCalledWith('user_registrations')
      expect(mockUpsert).toHaveBeenCalledWith({
        email: 'test@example.com',
        user_id: 'github-123',
        name: 'Test User',
        providers: ['github'],
        registered_at: '2025-01-01T00:00:00Z',
        installation_key: 'key-123',
        secret_key: null,
        has_completed_installation: false,
        subdomain: null,
        reserved_at: null,
        deployment_ready: null,
        last_health_check: null,
      })
    })

    it('should throw error when save fails', async () => {
      mockUpsert.mockResolvedValueOnce({
        error: { message: 'Database error' },
      })

      const registration: UserRegistration = {
        userId: 'github-123',
        email: 'test@example.com',
        name: 'Test User',
        providers: ['github'],
        registeredAt: '2025-01-01T00:00:00Z',
        installationKey: 'key-123',
        hasCompletedInstallation: false,
      }

      await expect(saveUserRegistration(registration)).rejects.toThrow(
        'Failed to save user registration: Database error',
      )
    })
  })

  describe('markInstallationComplete', () => {
    it('should mark installation complete with secret key', async () => {
      const mockUpdatedUser = {
        user_id: 'github-123',
        email: 'test@example.com',
        name: 'Test User',
        providers: ['github'],
        registered_at: '2025-01-01T00:00:00Z',
        installation_key: 'key-123',
        secret_key: 'new-secret',
        has_completed_installation: true,
      }

      mockUpdate.mockResolvedValueOnce({ error: null })
      mockSingle.mockResolvedValueOnce({ data: mockUpdatedUser, error: null })
      mockSelect.mockResolvedValueOnce({ data: [], error: null })

      const result = await markInstallationComplete('test@example.com', 'new-secret')

      expect(result.hasCompletedInstallation).toBe(true)
      expect(result.secretKey).toBe('new-secret')
      expect(mockUpdate).toHaveBeenCalledWith({
        has_completed_installation: true,
        secret_key: 'new-secret',
      })
    })

    it('should mark installation complete without secret key', async () => {
      const mockUpdatedUser = {
        user_id: 'github-123',
        email: 'test@example.com',
        name: 'Test User',
        providers: ['github'],
        registered_at: '2025-01-01T00:00:00Z',
        installation_key: 'key-123',
        has_completed_installation: true,
      }

      mockUpdate.mockResolvedValueOnce({ error: null })
      mockSingle.mockResolvedValueOnce({ data: mockUpdatedUser, error: null })
      mockSelect.mockResolvedValueOnce({ data: [], error: null })

      const result = await markInstallationComplete('test@example.com')

      expect(result.hasCompletedInstallation).toBe(true)
      expect(mockUpdate).toHaveBeenCalledWith({
        has_completed_installation: true,
      })
    })
  })

  describe('addOrUpdateIpRecord', () => {
    it('should add new IP record and return count', async () => {
      const ipRecord: IPRecord = {
        type: 'installation',
        ip: '1.2.3.4',
        configuredAt: '2025-01-01T00:00:00Z',
        dnsRecordIds: ['dns-1'],
      }

      mockUpsert.mockResolvedValueOnce({ error: null })
      mockSelect.mockResolvedValueOnce({
        data: null,
        error: null,
        count: 1,
      })

      const count = await addOrUpdateIpRecord('test@example.com', ipRecord)

      expect(count).toBe(1)
      expect(mockUpsert).toHaveBeenCalledWith(
        {
          user_email: 'test@example.com',
          type: 'installation',
          ip: '1.2.3.4',
          work_machine_name: null,
          configured_at: '2025-01-01T00:00:00Z',
          dns_record_ids: ['dns-1'],
        },
        {
          onConflict: 'user_email,type,work_machine_name',
        },
      )
    })

    it('should update existing IP record (upsert)', async () => {
      const ipRecord: IPRecord = {
        type: 'workmachine',
        ip: '1.2.3.5',
        workMachineName: 'dev1',
        configuredAt: '2025-01-01T00:00:00Z',
        dnsRecordIds: ['dns-2', 'dns-3'],
      }

      mockUpsert.mockResolvedValueOnce({ error: null })
      mockSelect.mockResolvedValueOnce({
        data: null,
        error: null,
        count: 2,
      })

      const count = await addOrUpdateIpRecord('test@example.com', ipRecord)

      expect(count).toBe(2)
      expect(mockUpsert).toHaveBeenCalledWith(
        {
          user_email: 'test@example.com',
          type: 'workmachine',
          ip: '1.2.3.5',
          work_machine_name: 'dev1',
          configured_at: '2025-01-01T00:00:00Z',
          dns_record_ids: ['dns-2', 'dns-3'],
        },
        {
          onConflict: 'user_email,type,work_machine_name',
        },
      )
    })

    it('should throw error when upsert fails', async () => {
      const ipRecord: IPRecord = {
        type: 'installation',
        ip: '1.2.3.4',
        configuredAt: '2025-01-01T00:00:00Z',
      }

      mockUpsert.mockResolvedValueOnce({
        error: { message: 'Constraint violation' },
      })

      await expect(addOrUpdateIpRecord('test@example.com', ipRecord)).rejects.toThrow(
        'Failed to add IP record: Constraint violation',
      )
    })
  })

  describe('isSubdomainAvailable', () => {
    it('should return false for reserved keywords', async () => {
      const result = await isSubdomainAvailable('api')
      expect(result).toBe(false)
    })

    it('should return false when subdomain is taken', async () => {
      mockSingle.mockResolvedValueOnce({
        data: { subdomain: 'test' },
        error: null,
      })

      const result = await isSubdomainAvailable('test')
      expect(result).toBe(false)
    })

    it('should return true when subdomain is available', async () => {
      mockSingle.mockResolvedValueOnce({
        data: null,
        error: { code: 'PGRST116' },
      })

      const result = await isSubdomainAvailable('myapp')
      expect(result).toBe(true)
    })
  })

  describe('reserveSubdomain', () => {
    it('should reserve subdomain successfully', async () => {
      const mockReservation = {
        subdomain: 'myapp',
        user_id: 'github-123',
        user_email: 'test@example.com',
        user_name: 'Test User',
        reserved_at: '2025-01-01T00:00:00Z',
        status: 'reserved',
      }

      mockSingle.mockResolvedValueOnce({ data: mockReservation, error: null })
      mockUpdate.mockResolvedValueOnce({ error: null })

      const result = await reserveSubdomain('myapp', 'github-123', 'test@example.com', 'Test User')

      expect(result.subdomain).toBe('myapp')
      expect(result.status).toBe('reserved')
      expect(mockInsert).toHaveBeenCalled()
      expect(mockUpdate).toHaveBeenCalled()
    })

    it('should throw error when subdomain already reserved', async () => {
      mockSingle.mockResolvedValueOnce({
        data: null,
        error: { code: '23505', message: 'Unique constraint violation' },
      })

      await expect(
        reserveSubdomain('taken', 'github-123', 'test@example.com', 'Test User'),
      ).rejects.toThrow('Subdomain is already reserved')
    })
  })

  describe('deleteIpRecords', () => {
    it('should delete all IP records and return DNS record IDs', async () => {
      const mockIpRecords = [
        { dns_record_ids: ['dns-1'] },
        { dns_record_ids: ['dns-2', 'dns-3'] },
        { dns_record_ids: null },
      ]

      mockSelect.mockResolvedValueOnce({ data: mockIpRecords, error: null })
      mockDelete.mockResolvedValueOnce({ error: null })

      const dnsIds = await deleteIpRecords('test@example.com')

      expect(dnsIds).toEqual(['dns-1', 'dns-2', 'dns-3'])
      expect(mockDelete).toHaveBeenCalled()
    })

    it('should return empty array when no IP records exist', async () => {
      mockSelect.mockResolvedValueOnce({ data: [], error: null })
      mockDelete.mockResolvedValueOnce({ error: null })

      const dnsIds = await deleteIpRecords('test@example.com')

      expect(dnsIds).toEqual([])
    })
  })

  describe('deleteDomainReservation', () => {
    it('should delete domain reservation successfully', async () => {
      mockDelete.mockResolvedValueOnce({ error: null })

      await expect(deleteDomainReservation('test@example.com')).resolves.toBeUndefined()
    })

    it('should throw error when delete fails', async () => {
      mockDelete.mockResolvedValueOnce({
        error: { message: 'Delete failed' },
      })

      await expect(deleteDomainReservation('test@example.com')).rejects.toThrow(
        'Failed to delete domain reservation: Delete failed',
      )
    })
  })

  describe('resetUserInstallation', () => {
    it('should reset installation successfully', async () => {
      mockUpdate.mockResolvedValueOnce({ error: null })

      await expect(resetUserInstallation('test@example.com')).resolves.toBeUndefined()

      expect(mockUpdate).toHaveBeenCalledWith({
        subdomain: null,
        reserved_at: null,
        secret_key: null,
        has_completed_installation: false,
        deployment_ready: false,
        last_health_check: null,
      })
    })

    it('should throw error when reset fails', async () => {
      mockUpdate.mockResolvedValueOnce({
        error: { message: 'Update failed' },
      })

      await expect(resetUserInstallation('test@example.com')).rejects.toThrow(
        'Failed to reset installation: Update failed',
      )
    })
  })
})
