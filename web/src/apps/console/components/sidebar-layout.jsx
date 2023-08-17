import { Link, Outlet } from '@remix-run/react';
import { SubHeader } from '~/components/organisms/sub-header';
import * as ActionList from '~/components/atoms/action-list';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';

const SidebarLayout = ({
  items = [],
  parentPath,
  headerTitle,
  headerActions = null,
}) => {
  const { activePath } = useActivePath({ parent: parentPath });
  return (
    <>
      <SubHeader title={headerTitle} actions={headerActions} />
      <div className="flex flex-row gap-10xl">
        <div className="w-[180px]">
          <ActionList.Root value={activePath} LinkComponent={Link}>
            {items.map((item) => (
              <ActionList.Button
                key={item.value}
                value={`/${item.value}`}
                href={item.value}
              >
                {item.label}
              </ActionList.Button>
            ))}
          </ActionList.Root>
        </div>
        <div className="flex-1 flex flex-col gap-6xl">
          <Outlet />
        </div>
      </div>
    </>
  );
};

export default SidebarLayout;
