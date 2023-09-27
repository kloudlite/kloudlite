import { Plus } from '@jengaicons/react';
import { Outlet, useOutletContext } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import SidebarLayout from '../components/sidebar-layout';
import { IAccountContext } from './_.$account';

const ContainerRegistry = () => {
  const rootContext = useOutletContext<IAccountContext>();
  const subNavAction = useSubNavData();
  return (
    <SidebarLayout
      headerActions={
        subNavAction.data &&
        subNavAction.data.show && (
          <Button
            content={subNavAction.data?.content}
            onClick={subNavAction.data.action}
            variant="primary"
            prefix={<Plus />}
          />
        )
      }
      navItems={[
        { label: 'Repos', value: 'repos' },
        { label: 'Access management', value: 'access-management' },
      ]}
      parentPath="/container-registry"
      headerTitle="Container registry"
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default ContainerRegistry;
