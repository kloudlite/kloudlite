import { Link } from '@remix-run/react';
import Tabs from '~/components/atoms/tabs';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';
import { ChevronLeft } from '@jengaicons/react';
import ScrollArea from '~/components/atoms/scroll-area';

export const CommonTabs = ({ tabs, baseurl, backButton = null }) => {
  const { activePath } = useActivePath({ parent: baseurl });

  return (
    <div className="flex flex-row items-center">
      {!!backButton && (
        <div className="flex flex-row items-center">
          <Link
            to={backButton.to}
            className="outline-none flex flex-row items-center gap-lg bodyMd-medium text-text-soft hover:text-text-default active:text-text-default py-lg cursor-pointer"
          >
            <ChevronLeft size={16} />
            {backButton.label}
          </Link>
          <span className="ml-4xl mr-2xl w-xs h-2xl bg-border-default" />
        </div>
      )}
      {/* <div className="-mx-3xl md:mx-0">
             
            </div> */}
      <ScrollArea
        blurfrom="from-white"
        rightblur={false}
        className="flex-1 -mr-2xl"
      >
        <Tabs.Root
          basePath={baseurl}
          value={`/${activePath.split('/')[1]}`}
          fitted
          LinkComponent={Link}
        >
          {tabs?.map(({ value, to, label }) => {
            return <Tabs.Tab {...{ value, to, label }} key={value} />;
          })}
        </Tabs.Root>
      </ScrollArea>
    </div>
  );
};
