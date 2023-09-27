import { Link } from '@remix-run/react';
import { ReactNode } from 'react';
import ActionList, { IActionItem } from '~/components/atoms/action-list';
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
        className="flex flex-row"
        onScroll={(e) => {
          console.log(e);
        }}
      >
        <div className="flex flex-col">
          <div className="sticky top-6xl flex flex-col">
            <div className="text-text-strong heading2xl min-h-[38px] pb-3xl pt-6xl">
              {headerTitle}
            </div>
            <div className="w-[180px] pt-3xl">
              <ActionList.Root value={activePath || ''} LinkComponent={Link}>
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
        <div
          className="flex flex-col flex-1 pl-10xl"
          onScroll={(e) => {
            console.log(e);
          }}
        >
          <div className="sticky z-10 bg-surface-basic-subdued top-6xl min-h-[38px]  py-6xl flex justify-end -mx-md px-md">
            {headerActions}
          </div>
          <div
            className="flex-1 flex flex-col gap-6xl"
            onScroll={(e) => {
              console.log(e);
            }}
          >
            {children}
          </div>
        </div>
      </div>
    </>
  );
};

export default SidebarLayout;
