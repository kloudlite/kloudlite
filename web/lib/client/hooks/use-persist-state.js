import { useEffect, useState } from 'react';

const usePersistState = (key, initialValue) => {
  if (!key) {
    throw new Error('key is required');
  }

  const [value, setValue] = useState(() => {
    if (typeof window === 'undefined') {
      return [initialValue, () => {}];
    }

    const storedValue = JSON.parse(
      localStorage.getItem('kl-persist-state') || '{}'
    );

    return storedValue[key] || initialValue;
    // return storedValue ? JSON.parse(storedValue) : initialValue;
  });

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    const storedValue = JSON.parse(
      localStorage.getItem('kl-persist-state') || '{}'
    );

    storedValue[key] = value;

    localStorage.setItem('kl-persist-state', JSON.stringify(storedValue));
  }, [key, value]);

  return [value, setValue];
};

export default usePersistState;
