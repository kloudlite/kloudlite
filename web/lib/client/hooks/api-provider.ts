import { useContext, createContext } from 'react';
// @ts-ignore
import { createRPCClient } from '@madhouselabs/madrpc';
import { LibApiType } from '../../server/gql/saved-queries';

export const APIContext = createContext(createRPCClient('/api'));

export const useAPIClient = () => useContext(APIContext);
export const useLibApi: () => LibApiType = useAPIClient;
