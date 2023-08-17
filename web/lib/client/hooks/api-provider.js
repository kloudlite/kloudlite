import { useContext, createContext, useState, useEffect } from 'react';
import { createRPCClient } from '@madhouselabs/madrpc';

// export const APIContext = React.createContext(createRPCClient('/api'));

export const APIContext = createContext(createRPCClient('/api'));

export const useAPIClient = () => useContext(APIContext);

export const useApiCall = ({ fn, data }) => {
  const [_data, setData] = useState();
  const [error, setError] = useState();
  const [isLoading, setIsLoading] = useState(true);
  useEffect(() => {
    (async () => {
      setIsLoading(true);
      try {
        const { data: __data, errors } = await fn();
        if (errors) {
          throw errors[0];
        }
        setData(__data);
      } catch (err) {
        setError(err.message);
        console.log(err);
      } finally {
        setIsLoading(false);
      }
    })();
  }, [data]);
  return { data: _data, error, isLoading };
};
