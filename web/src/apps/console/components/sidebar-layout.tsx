import { Link } from '@remix-run/react';
import { ReactNode } from 'react';
import ActionList, { IActionItem } from '~/components/atoms/action-list';
import ScrollArea from '~/components/atoms/scroll-area';
import Tabs from '~/components/atoms/tabs';
import { SubHeader } from '~/components/organisms/sub-header';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';

interface Item extends Omit<IActionItem, 'children'> {
  label: ReactNode;
}

interface ISidebarLayout {
  navItems: Item[];
  parentPath: string;
  headerTitle?: string;
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
      {!!headerTitle || !!headerActions ? (
        <SubHeader title={headerTitle} actions={headerActions} />
      ) : (
        <div className="pt-6xl" />
      )}
      <div className="flex flex-col md:flex-row">
        <div className="flex flex-col">
          <div className="flex flex-col">
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
        </div>
        {/* If overflow problem occurs in error page look here */}
        <div className="flex flex-col flex-1 md:pl-6xl">
          <div className="flex-1 flex flex-col gap-6xl">{children}</div>
        </div>
      </div>
    </>
  );
};

export default SidebarLayout;
