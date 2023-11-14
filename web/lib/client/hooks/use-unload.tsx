import { useEffect } from 'react';

const useUnload = ({ callback }: { callback: (e: Event) => void }) => {
  useEffect(() => {
    window.addEventListener('beforeunload', callback);
    return () => {
      window.removeEventListener('beforeunload', callback);
    };
  }, [callback]);
};

export default useUnload;
