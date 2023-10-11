import { Link } from '@remix-run/react';
import { ReactNode } from 'react';
import ActionList, { IActionItem } from '~/components/atoms/action-list';
import ScrollArea from '~/components/atoms/scroll-area';
import Tabs from '~/components/atoms/tabs';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';

interface Item extends Omit<IActionItem, 'children'> {
  label: ReactNode;
}

interface ISidebarLayout {
  navItems: Item[];
  parentPath: string;
  headerTitle: string;
  children: ReactNode;
  headerActions?: ReactNode;
}

const SidebarLayout = ({
  navItems = [],
  parentPath,
  headerTitle,
  headerActions,
  children = null,
}: ISidebarLayout) => {
  const { activePath } = useActivePath({ parent: parentPath });
  return (
    <>
      {/* <SubHeader title={headerTitle} actions={headerActions} /> */}
      <div
        className="flex flex-col md:flex-row"
        onScroll={(e) => {
          console.log(e);
        }}
      >
        <div className="flex flex-col">
          <div className="flex flex-col">
            <div className="text-text-strong heading2xl py-4xl md:py-6xl">
              <div className="min-h-[38px] flex flex-row items-center">
                {headerTitle}
              </div>
            </div>
            <div className="flex md:hidden pt-3xl pb-6xl">
              <ScrollArea
                blurfrom="from-white"
                rightblur={false}
                className="flex-1"
              >
                <Tabs.Root
                  value={activePath || ''}
                  LinkComponent={Link}
                  variant="filled"
                >
                  {navItems.map((item) => (
                    <Tabs.Tab
                      key={item.value}
                      value={`/${item.value}`}
                      to={item.value}
                      label={item.label}
                    />
                  ))}
                </Tabs.Root>
              </ScrollArea>
            </div>
            <div className="w-[180px] hidden md:flex">
              <ActionList.Root
                value={activePath || ''}
                LinkComponent={Link}
                className="w-full"
              >
                {navItems.map((item) => (
                  <ActionList.Item
                    key={item.value}
                    value={`/${item.value}`}
                    to={item.value}
                  >
                    {item.label}
                  </ActionList.Item>
                ))}
              </ActionList.Root>
            </div>
          </div>

          <div className="flex-1" />
        </div>
        <div className="flex flex-col flex-1 overflow-x-hidden md:pl-10xl">
          <div className="hidden bg-surface-basic-subdued top-6xl py-6xl flex-row gap-lg justify-end -mx-md px-md min-h-[38px] md:flex">
            {headerActions}
            {!headerActions && <span className="min-h-[38px]">&nbsp;</span>}
          </div>
          <div className="flex-1 flex flex-col gap-6xl">{children}</div>
        </div>
      </div>
    </>
  );
};

export default SidebarLayout;
