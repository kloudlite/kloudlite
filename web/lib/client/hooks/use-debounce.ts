import { useEffect } from 'react';

type IuseDebounce = (action: () => any, delay: number, dep: any[]) => void;

const useDebounce: IuseDebounce = (action, delay, dep = []) => {
  // State and setters for debounced value
  // const [debouncedValue, setDebouncedValue] = useState(value);
  useEffect(
    () => {
      let resp = () => {};
      // Update debounced value after delay
      const handler = setTimeout(() => {
        resp = action();
      }, delay);
      // Cancel the timeout if value changes (also on delay change or unmount)
      // This is how we prevent debounced value from updating if value is changed ...
      // .. within the delay period. Timeout gets cleared and restarted.
      return () => {
        clearTimeout(handler);
        if (typeof resp === 'function') {
          resp();
        }
      };
    },
    [delay, ...dep] // Only re-call effect if value or delay changes
  );
};

export default useDebounce;
