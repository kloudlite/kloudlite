'use client';

import { Settings, User, Shield } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { DashboardMode } from '@/types/dashboard';

interface AdminModeToggleProps {
  mode: DashboardMode;
  onModeChange: (mode: DashboardMode) => void;
}

export function AdminModeToggle({ mode, onModeChange }: AdminModeToggleProps) {
  return (
    <Button
      variant={mode === 'admin' ? 'default' : 'outline'}
      size="sm"
      onClick={() => onModeChange(mode === 'admin' ? 'user' : 'admin')}
      className="flex items-center gap-2 transition-colors duration-200"
    >
      {mode === 'admin' ? (
        <>
          <Shield className="h-4 w-4" />
          Admin View
        </>
      ) : (
        <>
          <Settings className="h-4 w-4" />
          Admin View
        </>
      )}
    </Button>
  );
}