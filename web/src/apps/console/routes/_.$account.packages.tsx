import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '../components/sidebar-layout';
import { IAccountContext } from './_.$account';

const ContainerRegistry = () => {
  const rootContext = useOutletContext<IAccountContext>();

  return (
    <SidebarLayout
      navItems={[
        { label: 'Container Repos', value: 'repos' },
        { label: 'Helm Repos', value: 'helm-repos' },
        { label: 'Access management', value: 'access-management' },
      ]}
      parentPath="/packages"
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default ContainerRegistry;
