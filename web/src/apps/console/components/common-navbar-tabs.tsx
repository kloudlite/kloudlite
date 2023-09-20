import { ChevronLeft } from '@jengaicons/react';
import { Link } from '@remix-run/react';
import { AnimatePresence, motion } from 'framer-motion';
import { useContext } from 'react';
import ScrollArea from '~/components/atoms/scroll-area';
import Tabs from '~/components/atoms/tabs';
import { BrandLogo } from '~/components/branding/brand-logo';
import { TopBarContext } from '~/components/organisms/top-bar';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';

interface CommonTabsProps {
  tabs: {
    label: string;
    to: string;
    value: string;
  }[];
  baseurl: any;
  backButton?: {
    to: string;
    label: string;
  } | null;
}

export const CommonTabs = ({
  tabs,
  baseurl,
  backButton = null,
}: CommonTabsProps) => {
  const { activePath } = useActivePath({ parent: baseurl });

  const context = useContext(TopBarContext);
  const { isSticked } = context || {};

  return (
    <div className="flex flex-row items-center">
      <AnimatePresence>
        {!!backButton && (
          <motion.div
            layoutId="back-button"
            initial={{ y: 0, width: 0, opacity: 0 }}
            exit={{ y: 0, width: 0, opacity: 0 }}
            animate={{ width: 'auto', opacity: 1 }}
            transition={{ duration: 0.2, type: 'spring', bounce: 0.1 }}
            className="flex flex-row items-center overflow-hidden"
          >
            <Link
              to={backButton.to}
              className="whitespace-nowrap outline-none flex flex-row items-center gap-lg bodyMd-medium text-text-soft hover:text-text-default active:text-text-default py-lg cursor-pointer"
            >
              <ChevronLeft size={16} />
              {backButton.label}
            </Link>
            <span className="ml-4xl mr-2xl w-xs h-2xl bg-border-default" />
          </motion.div>
        )}
      </AnimatePresence>

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
          {tabs.map(({ value, to, label }) => {
            return <Tabs.Tab {...{ value, to, label }} key={value} />;
          })}
        </Tabs.Root>
      </ScrollArea>

      <AnimatePresence>
        {!!isSticked && (
          <motion.div
            layoutId="small-logo"
            initial={{ y: 10, opacity: 0 }}
            exit={{ y: 10, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            transition={{ duration: 0.2, type: 'spring', bounce: 0.1 }}
            className="flex flex-row items-center overflow-hidden"
          >
            <div className="flex justify-center items-center pl-2xl">
              <BrandLogo size={18} detailed />
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};
