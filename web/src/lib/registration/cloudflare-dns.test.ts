import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  createDnsRecord,
  updateDnsRecord,
  deleteDnsRecord,
  getDnsRecord,
  createInstallationDnsRecords,
  createWorkmachineDnsRecords,
  updateDnsRecords,
  deleteDnsRecords
} from './cloudflare-dns'

// Mock global fetch
global.fetch = vi.fn()

describe('Cloudflare DNS Service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('createDnsRecord', () => {
    it('should create DNS A record successfully', async () => {
      const mockResponse = {
        success: true,
        errors: [],
        messages: [],
        result: {
          id: 'dns-record-123',
          type: 'A',
          name: 'test.test.dev',
          content: '1.2.3.4',
          proxied: false,
          ttl: 120
        }
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const recordId = await createDnsRecord('test.test.dev', '1.2.3.4')

      expect(recordId).toBe('dns-record-123')
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/dns_records'),
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Authorization': expect.stringContaining('Bearer'),
            'Content-Type': 'application/json'
          }),
          body: expect.stringContaining('"type":"A"')
        })
      )
    })

    it('should return null when API call fails', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        text: async () => 'API Error'
      })

      const recordId = await createDnsRecord('test.test.dev', '1.2.3.4')

      expect(recordId).toBeNull()
    })

    it('should return null when response indicates failure', async () => {
      const mockResponse = {
        success: false,
        errors: [{ code: 1004, message: 'DNS validation failed' }],
        messages: [],
        result: null
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const recordId = await createDnsRecord('invalid.test.dev', '1.2.3.4')

      expect(recordId).toBeNull()
    })

    it('should create proxied DNS record', async () => {
      const mockResponse = {
        success: true,
        errors: [],
        messages: [],
        result: {
          id: 'dns-record-456',
          type: 'A',
          name: 'cdn.test.dev',
          content: '1.2.3.4',
          proxied: true,
          ttl: 120
        }
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const recordId = await createDnsRecord('cdn.test.dev', '1.2.3.4', true)

      expect(recordId).toBe('dns-record-456')
    })
  })

  describe('updateDnsRecord', () => {
    it('should update DNS record successfully', async () => {
      const mockResponse = {
        success: true,
        errors: [],
        messages: [],
        result: {
          id: 'dns-record-123',
          type: 'A',
          name: 'test.test.dev',
          content: '5.6.7.8',
          proxied: false,
          ttl: 120
        }
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const success = await updateDnsRecord('dns-record-123', 'test.test.dev', '5.6.7.8')

      expect(success).toBe(true)
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/dns_records/dns-record-123'),
        expect.objectContaining({
          method: 'PATCH'
        })
      )
    })

    it('should return false when update fails', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        text: async () => 'Update failed'
      })

      const success = await updateDnsRecord('dns-record-123', 'test.test.dev', '5.6.7.8')

      expect(success).toBe(false)
    })
  })

  describe('deleteDnsRecord', () => {
    it('should delete DNS record successfully', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        text: async () => 'Success'
      })

      const success = await deleteDnsRecord('dns-record-123')

      expect(success).toBe(true)
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/dns_records/dns-record-123'),
        expect.objectContaining({
          method: 'DELETE'
        })
      )
    })

    it('should return true when record not found (404)', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 404,
        text: async () => 'Not found'
      })

      const success = await deleteDnsRecord('dns-record-123')

      expect(success).toBe(true)
    })

    it('should return false on other errors', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500,
        text: async () => 'Server error'
      })

      const success = await deleteDnsRecord('dns-record-123')

      expect(success).toBe(false)
    })
  })

  describe('getDnsRecord', () => {
    it('should get DNS record successfully', async () => {
      const mockResponse = {
        success: true,
        errors: [],
        messages: [],
        result: [{
          id: 'dns-record-123',
          type: 'A',
          name: 'test.test.dev',
          content: '1.2.3.4',
          proxied: false,
          ttl: 120
        }]
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const record = await getDnsRecord('test.test.dev')

      expect(record).toEqual({
        id: 'dns-record-123',
        type: 'A',
        name: 'test.test.dev',
        content: '1.2.3.4',
        proxied: false,
        ttl: 120
      })
    })

    it('should return null when record not found', async () => {
      const mockResponse = {
        success: true,
        errors: [],
        messages: [],
        result: []
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const record = await getDnsRecord('nonexistent.test.dev')

      expect(record).toBeNull()
    })
  })

  describe('createInstallationDnsRecords', () => {
    it('should create installation DNS record', async () => {
      const mockResponse = {
        success: true,
        errors: [],
        messages: [],
        result: {
          id: 'dns-record-installation',
          type: 'A',
          name: 'myapp.test.dev',
          content: '1.2.3.4',
          proxied: false,
          ttl: 120
        }
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const recordIds = await createInstallationDnsRecords('myapp', '1.2.3.4')

      expect(recordIds).toEqual(['dns-record-installation'])
      expect(recordIds.length).toBe(1)
    })

    it('should return empty array if creation fails', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        text: async () => 'Failed'
      })

      const recordIds = await createInstallationDnsRecords('myapp', '1.2.3.4')

      expect(recordIds).toEqual([])
    })
  })

  describe('createWorkmachineDnsRecords', () => {
    it('should create workmachine DNS records (exact and wildcard)', async () => {
      const exactMockResponse = {
        success: true,
        errors: [],
        messages: [],
        result: {
          id: 'dns-record-exact',
          type: 'A',
          name: 'dev1.myapp.test.dev',
          content: '1.2.3.5',
          proxied: false,
          ttl: 120
        }
      }

      const wildcardMockResponse = {
        success: true,
        errors: [],
        messages: [],
        result: {
          id: 'dns-record-wildcard',
          type: 'A',
          name: '*.dev1.myapp.test.dev',
          content: '1.2.3.5',
          proxied: false,
          ttl: 120
        }
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => exactMockResponse
      })
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => wildcardMockResponse
      })

      const recordIds = await createWorkmachineDnsRecords('dev1', 'myapp', '1.2.3.5')

      expect(recordIds).toEqual(['dns-record-exact', 'dns-record-wildcard'])
      expect(recordIds.length).toBe(2)
    })

    it('should return partial records if one fails', async () => {
      const exactMockResponse = {
        success: true,
        errors: [],
        messages: [],
        result: {
          id: 'dns-record-exact',
          type: 'A',
          name: 'dev1.myapp.test.dev',
          content: '1.2.3.5',
          proxied: false,
          ttl: 120
        }
      }

      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => exactMockResponse
      })
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        text: async () => 'Failed'
      })

      const recordIds = await createWorkmachineDnsRecords('dev1', 'myapp', '1.2.3.5')

      expect(recordIds).toEqual(['dns-record-exact'])
      expect(recordIds.length).toBe(1)
    })
  })

  describe('updateDnsRecords', () => {
    it('should update multiple DNS records successfully', async () => {
      ;(global.fetch as any).mockResolvedValue({
        ok: true,
        json: async () => ({
          success: true,
          errors: [],
          messages: [],
          result: {}
        })
      })

      const success = await updateDnsRecords(
        ['dns-1', 'dns-2'],
        'test.test.dev',
        '5.6.7.8'
      )

      expect(success).toBe(true)
      expect(global.fetch).toHaveBeenCalledTimes(2)
    })

    it('should return false if any update fails', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          errors: [],
          messages: [],
          result: {}
        })
      })
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        text: async () => 'Failed'
      })

      const success = await updateDnsRecords(
        ['dns-1', 'dns-2'],
        'test.test.dev',
        '5.6.7.8'
      )

      expect(success).toBe(false)
    })
  })

  describe('deleteDnsRecords', () => {
    it('should delete multiple DNS records successfully', async () => {
      ;(global.fetch as any).mockResolvedValue({
        ok: true,
        text: async () => 'Success'
      })

      const success = await deleteDnsRecords(['dns-1', 'dns-2', 'dns-3'])

      expect(success).toBe(true)
      expect(global.fetch).toHaveBeenCalledTimes(3)
    })

    it('should return false if any deletion fails', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        text: async () => 'Success'
      })
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500,
        text: async () => 'Server error'
      })

      const success = await deleteDnsRecords(['dns-1', 'dns-2'])

      expect(success).toBe(false)
    })
  })
})
