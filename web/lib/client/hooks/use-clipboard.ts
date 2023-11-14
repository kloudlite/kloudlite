import { useEffect, useState } from 'react';

const useClipboard = ({
  onSuccess,
}: {
  onSuccess?: (data: string) => void;
}) => {
  const [data, setData] = useState<string | null>(null);
  useEffect(() => {
    if (data) {
      (async () => {
        await navigator.clipboard.writeText(data);
        if (onSuccess) {
          onSuccess(data);
        }
        setData(null);
      })();
    }
  }, [data]);
  return { copy: setData };
};

export default useClipboard;
