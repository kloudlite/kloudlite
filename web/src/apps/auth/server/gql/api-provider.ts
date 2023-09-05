import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { AuthApiType } from './saved-queries';

export const useAuthApi: () => AuthApiType = useAPIClient;
