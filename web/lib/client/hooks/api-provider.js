import { useContext, createContext } from 'react';
import { createRPCClient } from '@madhouselabs/madrpc';

// export const APIContext = React.createContext(createRPCClient('/api'));

export const APIContext = createContext(createRPCClient('/api'));

export const useAPIClient = () => useContext(APIContext);
