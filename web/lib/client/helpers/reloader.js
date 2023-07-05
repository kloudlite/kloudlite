import { useLocation, useNavigate } from 'react-router';

export const useReload = () => {
  const location = useLocation();
  const navigate = useNavigate();
  return () => {
    navigate(location.pathname + location.search, { replace: true });
  };
};
