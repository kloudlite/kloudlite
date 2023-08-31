import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { IGQLMethodsAuth } from './saved-queries';

export const useAuthApi = (): IGQLMethodsAuth => {
  return useAPIClient();
};
