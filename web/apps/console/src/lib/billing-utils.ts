const currencySymbols: Record<string, string> = {
  INR: '₹',
  USD: '$',
  EUR: '€',
  GBP: '£',
}

export function getCurrencySymbol(currency: string = 'INR'): string {
  return currencySymbols[currency.toUpperCase()] ?? currency
}

export function formatCurrency(amountInPaise: number, currency: string = 'INR'): string {
  return `${getCurrencySymbol(currency)}${(amountInPaise / 100).toFixed(2)}`
}
