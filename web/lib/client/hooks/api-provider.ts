import { useContext, createContext } from 'react';
// @ts-ignore
// import { createRPCClient } from '@madhouselabs/madrpc';
import { LibApiType } from '../../server/gql/saved-queries';
import { createRPCClient } from '../../server/helpers/rpc';

export const APIContext = createContext(createRPCClient('/api'));

export const useAPIClient = () => useContext(APIContext);
export const useLibApi: () => LibApiType = useAPIClient;

export const useCachedClient = () => {};
