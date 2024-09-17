//@ts-ignore
import { cn } from 'kl-design-system/utils';
import React, { ReactNode } from 'react';

const Wrapper = ({
    children,
    className,
}: {
    children: ReactNode;
    className?: string;
}) => {
    return (
        <div
            className={cn(
                'lg:m-auto lg:max-w-[896px] w-full px-3xl md:px-5xl lg:px-8xl xl:px-11xl 2xl:px-12xl xl:max-w-[1024px] 2xl:max-w-[1120px] 3xl:min-w-[1408px] lg:box-content',
                className
            )}
        >
            {children}
        </div>
    );
};

export default Wrapper;
