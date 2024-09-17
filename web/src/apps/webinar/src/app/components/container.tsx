//@ts-ignore
import { cn } from 'kl-design-system/utils';
import { ReactNode } from 'react';
import Header from './header';

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
                    'flex flex-1 flex-col md:items-center self-stretch justify-center px-3xl py-5xl md:py-9xl'
                )}
            >
                {children}
            </div>
            {/* <Footer /> */}
        </div>
    );
};

export default Container;
