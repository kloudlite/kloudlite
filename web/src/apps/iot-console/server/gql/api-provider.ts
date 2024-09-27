import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { ConsoleApiType } from './saved-queries';

export const useIotConsoleApi: () => ConsoleApiType = useAPIClient;
