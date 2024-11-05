import { useEffect, useState } from 'react';
declare global {
  interface Window {
    env: {
      [key: string]: any;
    };
  }
}

const useEnv = () => {
  const [env, setEnv] = useState<any>({});
  useEffect(() => {
    setEnv(window);
  });
  return env;
};

export default useEnv;
