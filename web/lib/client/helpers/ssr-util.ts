import { useEffect, useState } from 'react';
import { ChildrenProps } from '@kloudlite/design-system/types';

export const SafeHydrate = ({ children }: ChildrenProps) => {
  const [hasMounted, setHasMounted] = useState(false);

  useEffect(() => {
    setHasMounted(true);
  }, []);
  if (!hasMounted) {
    return null;
  }
  return children;
};
