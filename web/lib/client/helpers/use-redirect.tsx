import { useNavigate } from '@remix-run/react';

export const useExternalRedirect = () => {
  const navigate = useNavigate();
  return (url = '/') => {
    navigate(`/redirect?url=${url}`);
  };
};
