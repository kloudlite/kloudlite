import { Link } from '@remix-run/react';
import { ReactNode } from 'react';
import ActionList, { IActionItem } from '~/components/atoms/action-list';
import { SubHeader } from '~/components/organisms/sub-header';
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
      <SubHeader title={headerTitle} actions={headerActions} />
      <div className="flex flex-row gap-10xl">
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
        <div className="flex-1 flex flex-col gap-6xl">{children}</div>
      </div>
    </>
  );
};

export default SidebarLayout;
