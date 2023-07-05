import { useEffect } from 'react';

const useMenuHandler = ({ setOpen, contentRef }) => {
  useEffect(() => {
    const clickListener = (e) => {
      if (contentRef.current && !contentRef.current.contains(e.target)) {
        e.stopPropagation();
        setOpen(false);
      }
    };
    const keyDownListener = (e) => {
      if (e.code === 'Escape') {
        e.stopPropagation();
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', clickListener);
    document.addEventListener('keydown', keyDownListener);
    return () => {
      document.removeEventListener('mousedown', clickListener);
      document.removeEventListener('keydown', keyDownListener);
    };
  }, [setOpen, contentRef]);
};

export default useMenuHandler;
