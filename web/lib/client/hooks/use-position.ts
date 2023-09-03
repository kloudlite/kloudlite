import { RefObject, useEffect, useState } from 'react';

export const useSticky = (elementRef: RefObject<HTMLElement>, topLimit = 0) => {
  const [isStickey, setIsSticky] = useState(false);

  useEffect(() => {
    const getScroll = () => {
      if (elementRef && elementRef.current) {
        const { top } = elementRef.current.getBoundingClientRect();
        if (top < topLimit) {
          setIsSticky(true);
        } else {
          setIsSticky(false);
        }
      }
    };
    document.addEventListener('scroll', getScroll);
    return () => {
      document.removeEventListener('scroll', getScroll);
    };
  }, [elementRef, topLimit]);

  return isStickey;
};
