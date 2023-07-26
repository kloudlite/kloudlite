import { useContext, createContext } from 'react';
import { createRPCClient } from '@madhouselabs/madrpc';

export const APIContext = createContext(
  createRPCClient('/api/')
  // createRPCClient('/api/')
);

export const useAPIClient = () => useContext(APIContext);
