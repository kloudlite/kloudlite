import {ReactNode, useState} from 'react';
import Tooltip from '~/components/atoms/tooltip';
import {cn, titleCase} from '~/components/utils';
import {CopyrightFill, CopySimple} from "@jengaicons/react";
import useClipboard from "~/lib/client/hooks/use-clipboard";
import {toast} from "~/components/molecule/toast";

interface IBase {
    className?: string;
    action?: ReactNode;
}

const BaseStyle = 'flex flex-row items-center gap-xl';

const ListSecondary = ({
                           className,
                           action,
                           title,
                           avatar,
                           subtitle,
                       }: {
    className?: string;
    action?: ReactNode;
    title?: ReactNode;
    subtitle?: ReactNode;
    avatar?: ReactNode;
}) => {
    return (
        <div className={cn(BaseStyle, className)}>
            <div className="flex flex-row items-center gap-xl flex-1 truncate">
                {avatar}
                <div className="flex flex-col gap-sm flex-1 truncate">
                    {title && (
                        <div className="bodySm truncate text-text-soft pulsable">
                            {title}
                        </div>
                    )}

                    {subtitle && (
                        <div className="bodyMd-medium truncate pulsable">{subtitle}</div>
                    )}
                </div>
            </div>
            {action}
        </div>
    );
};

const ListBody = ({
                      data,
                      className = '',
                      action,
                  }: {
    data: ReactNode;
} & IBase) => {
    return (
        <div
            className={cn('bodyMd text-text-strong truncate', BaseStyle, className)}
        >
            <div className="flex-1 truncate pulsable">{data}</div>
            {action}
        </div>
    );
};

const ListItem = ({
                      data,
                      subtitle,
                      className = '',
                      action,
                  }: {
    data?: ReactNode;
    subtitle?: ReactNode;
} & IBase) => {
    return (
        <div className={cn(BaseStyle, className)}>
            <div className="flex flex-col flex-1 truncate">
                {data && (
                    <div className="flex-1 bodyMd-medium text-text-strong truncate pulsable">
                        {data}
                    </div>
                )}
                {subtitle && (
                    <div className="pulsable bodyMd text-text-soft truncate">
                        {subtitle}
                    </div>
                )}
            </div>
            {action}
        </div>
    );
};

const ListTitle = ({
                       className,
                       action,
                       title,
                       avatar,
                       subtitle,
                   }: {
    className?: string;
    action?: ReactNode;
    title?: ReactNode;
    subtitle?: ReactNode;
    avatar?: ReactNode;
}) => {
    return (
        <div className={cn(BaseStyle, className)}>
            <div className="flex flex-row items-center gap-xl flex-1 truncate">
                {avatar}
                <div className="flex flex-col gap-sm flex-1 truncate">
                    {title && (
                        <div className="bodyMd-semibold text-text-default truncate pulsable">
                            <Tooltip.Root
                                className="!w-fit !max-w-fit"
                                side="top"
                                content={<div className="bodySm text-text-strong">{title}</div>}
                            >
                                <span className="w-fit">{title}</span>
                            </Tooltip.Root>
                        </div>
                    )}

                    {subtitle && (
                        <div className="bodySm text-text-soft truncate pulsable">
                            {subtitle}
                        </div>
                    )}
                </div>
            </div>
            {action}
        </div>
    );
};

const ListDomainItem = ({
                        data,
                        value,
                    }: {
    data: ReactNode;
    value: string;
}) => {
    const [_, setCopyIcon] = useState(<CopySimple/>);
    const {copy} = useClipboard({
        onSuccess: () => {
            setTimeout(() => {
                setCopyIcon(<CopyrightFill/>);
                toast.success(`${titleCase("domain name")} copied successfully`);
            }, 1000);
            // toast.success('Copied to clipboard');
        },
    });

    return (
        <div
            onClick={(event) => {
                event.preventDefault()
                copy(value);
            }}
            className="flex flex-row gap-md items-center select-none group cursor-pointer"
        >
            <div className="flex flex-col flex-1">
                {data && (
                    <div className="bodyMd-medium text-text-soft truncate pulsable">
                        {data}
                    </div>
                )}
            </div>
            <span className="invisible group-hover:visible">
                <CopySimple size={10}/>
            </span>
        </div>
    );
};

const listFlex = ({key}: { key: string }) => ({
    key,
    className: 'basis-full',
    render: () => <div/>,
});

const listClass = {
    title: 'w-[180px] min-w-[180px] max-w-[180px] mr-2xl',
    author: 'w-[180px] min-w-[180px] max-w-[180px]',
};
export {ListBody, ListItem, ListTitle, ListSecondary, listFlex, ListDomainItem, listClass};
