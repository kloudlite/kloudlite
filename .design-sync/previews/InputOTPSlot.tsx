import * as React from 'react'
import { InputOTP, InputOTPGroup, InputOTPSlot } from '@kloudlite/ui'

export const FilledSlots = () => (
  <InputOTP maxLength={6} value="482913">
    <InputOTPGroup>
      <InputOTPSlot index={0} />
      <InputOTPSlot index={1} />
      <InputOTPSlot index={2} />
      <InputOTPSlot index={3} />
      <InputOTPSlot index={4} />
      <InputOTPSlot index={5} />
    </InputOTPGroup>
  </InputOTP>
)

export const EmptySlots = () => (
  <InputOTP maxLength={6} value="">
    <InputOTPGroup>
      <InputOTPSlot index={0} />
      <InputOTPSlot index={1} />
      <InputOTPSlot index={2} />
      <InputOTPSlot index={3} />
      <InputOTPSlot index={4} />
      <InputOTPSlot index={5} />
    </InputOTPGroup>
  </InputOTP>
)

export const LargeSlots = () => (
  <InputOTP maxLength={4} value="9051">
    <InputOTPGroup>
      <InputOTPSlot index={0} className="h-12 w-12 text-base" />
      <InputOTPSlot index={1} className="h-12 w-12 text-base" />
      <InputOTPSlot index={2} className="h-12 w-12 text-base" />
      <InputOTPSlot index={3} className="h-12 w-12 text-base" />
    </InputOTPGroup>
  </InputOTP>
)

export const DisabledSlots = () => (
  <InputOTP maxLength={6} value="613827" disabled>
    <InputOTPGroup>
      <InputOTPSlot index={0} />
      <InputOTPSlot index={1} />
      <InputOTPSlot index={2} />
      <InputOTPSlot index={3} />
      <InputOTPSlot index={4} />
      <InputOTPSlot index={5} />
    </InputOTPGroup>
  </InputOTP>
)
