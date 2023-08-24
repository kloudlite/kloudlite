import { useEffect, useState } from 'react';
// import logger from '../helpers/log';

export const useSticky = (elementRef, topLimit = 0) => {
  const [isStickey, setIsSticky] = useState(false);

  useEffect(() => {
    const getScroll = () => {
      if (elementRef && elementRef.current) {
        const { top } = elementRef.current.getBoundingClientRect();
        // if (log) {
        //   logger.log(top, topLimit);
        // }

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
