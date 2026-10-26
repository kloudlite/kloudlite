import * as React from 'react'
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSeparator,
  InputOTPSlot,
  Label,
} from '@kloudlite/ui'

export const SixDigitVerification = () => (
  <div className="space-y-2">
    <Label htmlFor="otp-signin">Verification code</Label>
    <InputOTP id="otp-signin" maxLength={6} value="482913">
      <InputOTPGroup>
        <InputOTPSlot index={0} />
        <InputOTPSlot index={1} />
        <InputOTPSlot index={2} />
        <InputOTPSlot index={3} />
        <InputOTPSlot index={4} />
        <InputOTPSlot index={5} />
      </InputOTPGroup>
    </InputOTP>
  </div>
)

export const GroupedWithSeparator = () => (
  <InputOTP maxLength={6} value="204817">
    <InputOTPGroup>
      <InputOTPSlot index={0} />
      <InputOTPSlot index={1} />
      <InputOTPSlot index={2} />
    </InputOTPGroup>
    <InputOTPSeparator />
    <InputOTPGroup>
      <InputOTPSlot index={3} />
      <InputOTPSlot index={4} />
      <InputOTPSlot index={5} />
    </InputOTPGroup>
  </InputOTP>
)

export const PartiallyFilled = () => (
  <InputOTP maxLength={6} value="61">
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

export const Disabled = () => (
  <InputOTP maxLength={6} value="482913" disabled>
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

export const TwoFactorStep = () => (
  <div className="w-96 space-y-3 rounded-md border border-border p-6">
    <div className="space-y-1">
      <p className="text-sm font-semibold">Two-factor authentication</p>
      <p className="text-sm text-muted-foreground">
        Enter the 6-digit code from your authenticator app to continue to the console.
      </p>
    </div>
    <InputOTP maxLength={6} value="739204">
      <InputOTPGroup>
        <InputOTPSlot index={0} />
        <InputOTPSlot index={1} />
        <InputOTPSlot index={2} />
      </InputOTPGroup>
      <InputOTPSeparator />
      <InputOTPGroup>
        <InputOTPSlot index={3} />
        <InputOTPSlot index={4} />
        <InputOTPSlot index={5} />
      </InputOTPGroup>
    </InputOTP>
  </div>
)
