import { Outlet } from '@remix-run/react';
import SidebarLayout from '../components/sidebar-layout';

const ContainerRegistry = () => {
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        { label: 'Access management', value: 'access-management' },
      ]}
      parentPath="/container-registry"
      headerTitle="Container registry"
    >
      <Outlet />
    </SidebarLayout>
  );
};

export default ContainerRegistry;
