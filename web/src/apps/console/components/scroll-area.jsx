import { useState } from 'react';
import { cn } from '~/components/utils';

const ScrollArea = ({ children, className }) => {
  const [isScrolled, setIsScrolled] = useState(false);
  const handleScroll = ({ target }) => {
    setIsScrolled(target.scrollLeft > 0);
  };
  return (
    <div className={cn('w-0 relative', className)}>
      {isScrolled && (
        <div className="bg-gradient-to-r from-surface-basic-subdued to-transparent absolute h-full w-2xl -left-[3px] top-0 " />
      )}
      <div
        className="no-scrollbar overflow-x-scroll flex flex-row py-[3px] pl-[3px] -ml-[3px]"
        onScroll={handleScroll}
      >
        {children}
        <div className="min-w-[16px]" />
      </div>
      <div className="bg-gradient-to-l from-surface-basic-subdued to-transparent absolute h-full w-2xl right-0 top-0" />
    </div>
  );
};

export default ScrollArea;
