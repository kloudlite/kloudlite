import { useContext, createContext } from 'react';
import { createRPCClient } from '@madhouselabs/madrpc';

export const APIContext = createContext(
  createRPCClient('https://auth.kloudlite.io/api/')
);

export const useAPIClient = () => useContext(APIContext);
