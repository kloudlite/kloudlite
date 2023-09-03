import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { GQLServerHandler } from './saved-queries';

export const useAuthApi = () => {
  const dummyCtx: any = {};
  const handler = GQLServerHandler(dummyCtx);

  const api: typeof handler = useAPIClient();

  return api;
};
