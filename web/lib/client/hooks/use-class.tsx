import { useEffect, useState } from 'react';

const useClass = ({ elementClass }: { elementClass: string }) => {
  const [tempClass, setTempClass] = useState('');
  const [tempElementClass, setTempElementClass] = useState(elementClass);

  useEffect(() => {
    const element = document.querySelector(`.${tempElementClass}`)?.classList;
    if (tempClass) {
      element?.add(tempClass);
    }
  }, [tempClass, tempElementClass]);

  const removeClass = (className: string) => {
    const element = document.querySelector(`.${tempElementClass}`)?.classList;
    if (className) {
      element?.remove(className);
    }
    setTempClass('');
  };

  return {
    className: tempClass,
    elementClass: tempElementClass,
    setClassName: setTempClass,
    removeClassName: removeClass,
    setElementClass: setTempElementClass,
  };
};

export default useClass;
