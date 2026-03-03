export interface RazorpayOrder {
  id: string
  amount: number
  currency: string
  status: string
  receipt: string
  notes: Record<string, string>
}
