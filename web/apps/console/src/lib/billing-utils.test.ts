import { describe, it, expect } from 'vitest'
import { getCurrencySymbol, formatCurrency } from './billing-utils'

describe('billing-utils', () => {
  describe('getCurrencySymbol', () => {
    it('should return ₹ for INR', () => {
      expect(getCurrencySymbol('INR')).toBe('₹')
    })

    it('should return $ for USD', () => {
      expect(getCurrencySymbol('USD')).toBe('$')
    })

    it('should return € for EUR', () => {
      expect(getCurrencySymbol('EUR')).toBe('€')
    })

    it('should return £ for GBP', () => {
      expect(getCurrencySymbol('GBP')).toBe('£')
    })

    it('should be case-insensitive', () => {
      expect(getCurrencySymbol('inr')).toBe('₹')
      expect(getCurrencySymbol('usd')).toBe('$')
      expect(getCurrencySymbol('Eur')).toBe('€')
    })

    it('should default to USD when no currency is provided', () => {
      expect(getCurrencySymbol()).toBe('$')
    })

    it('should return the currency code itself for unknown currencies', () => {
      expect(getCurrencySymbol('JPY')).toBe('JPY')
      expect(getCurrencySymbol('AUD')).toBe('AUD')
    })
  })

  describe('formatCurrency', () => {
    it('should format amount in cents to USD by default', () => {
      expect(formatCurrency(290000)).toBe('$2900.00')
    })

    it('should format small amounts correctly', () => {
      expect(formatCurrency(100)).toBe('$1.00')
      expect(formatCurrency(50)).toBe('$0.50')
      expect(formatCurrency(1)).toBe('$0.01')
    })

    it('should format zero amount', () => {
      expect(formatCurrency(0)).toBe('$0.00')
    })

    it('should respect the currency parameter', () => {
      expect(formatCurrency(10000, 'USD')).toBe('$100.00')
      expect(formatCurrency(10000, 'EUR')).toBe('€100.00')
      expect(formatCurrency(10000, 'GBP')).toBe('£100.00')
    })

    it('should default to USD when no currency is provided', () => {
      expect(formatCurrency(2900)).toBe('$29.00')
    })

    it('should handle unknown currencies', () => {
      expect(formatCurrency(10000, 'JPY')).toBe('JPY100.00')
    })
  })
})
