import { useEffect, useState } from 'react';

const useScroll = (element: HTMLElement | null, topLimit = 0) => {
  const [reached, setReached] = useState(false);

  useEffect(() => {
    const scrollEvent = () => {
      if (element) {
        const { top } = element.getBoundingClientRect();
        if (top < topLimit) {
          setReached(true);
        } else {
          setReached(false);
        }
      }
    };
    document.addEventListener('scroll', scrollEvent);

    return () => {
      document.removeEventListener('scroll', scrollEvent);
    };
  }, [element, topLimit]);

  return reached;
};

export default useScroll;
