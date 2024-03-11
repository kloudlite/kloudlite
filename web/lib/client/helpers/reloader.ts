import { useRevalidator } from '@remix-run/react';

export const useReload = () => {
  const revalidator = useRevalidator();

  return () => {
    revalidator.revalidate();
  };
};
