import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { IGQLMethodsConsole } from './saved-queries';

export const useConsoleApi = (): IGQLMethodsConsole => {
  return useAPIClient();
};
