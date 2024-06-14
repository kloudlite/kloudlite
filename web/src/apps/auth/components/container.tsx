import { ReactNode } from 'react';
import { cn } from '~/components/utils';
import Header from './header';
import Footer from './footer';

interface IContainer {
  children: ReactNode;
  headerExtra?: ReactNode;
}

const Container = ({ children, headerExtra }: IContainer) => {
  return (
    <div className="flex flex-col h-full">
      <Header headerExtra={headerExtra} />
      <div
        className={cn(
          'flex flex-1 flex-col md:items-center self-stretch justify-center px-3xl py-9xl'
        )}
      >
        {children}
      </div>
      <Footer />
    </div>
  );
};

export default Container;
