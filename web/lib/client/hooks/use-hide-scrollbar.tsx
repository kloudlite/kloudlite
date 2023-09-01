import { useEffect } from 'react';

const useHideScrollBar = (hide = true) => {
  useEffect(() => {
    if (window?.document?.children[0] && hide)
      // @ts-ignore
      window.document.children[0].style = `overflow-y:hidden`;
    return () => {
      if (window?.document?.children[0])
        // @ts-ignore
        window.document.children[0].style = `overflow-y:auto`;
    };
  }, [hide]);
};

export default useHideScrollBar;
