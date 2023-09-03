import { useLocation, useNavigate } from '@remix-run/react';

export const useReload = () => {
  const location = useLocation();
  const navigate = useNavigate();
  return () => {
    navigate(location.pathname + location.search, { replace: true });
  };
};
