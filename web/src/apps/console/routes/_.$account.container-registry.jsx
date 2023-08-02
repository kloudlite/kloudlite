import { Outlet } from '@remix-run/react';
import { SubHeader } from '~/components/organisms/sub-header';
import * as ActionList from '~/components/atoms/action-list';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';

const ContainerRegistry = () => {
  const { activePath } = useActivePath({ parent: '/container-registry' });

  return (
    <>
      <SubHeader title="Container registry" />
      <div className="flex flex-row gap-10xl">
        <div className="w-[180px]">
          <ActionList.ActionRoot value={activePath}>
            <ActionList.ActionButton value="/general" href="general">
              General
            </ActionList.ActionButton>
            <ActionList.ActionButton
              value="/access-management"
              href="access-management"
            >
              Access management
            </ActionList.ActionButton>
          </ActionList.ActionRoot>
        </div>
        <div className="flex-1 flex flex-col gap-6xl">
          <Outlet />
        </div>
      </div>
    </>
  );
};

export default ContainerRegistry;
