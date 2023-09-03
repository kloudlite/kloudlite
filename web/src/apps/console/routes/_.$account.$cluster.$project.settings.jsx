import { Outlet } from '@remix-run/react';
import SidebarLayout from '../components/sidebar-layout';

const Settings = () => {
  return (
    <SidebarLayout
      navItems={[{ label: 'Access management', value: 'access-management' }]}
      parentPath="/settings"
      headerTitle="Settings"
    >
      <Outlet />
    </SidebarLayout>
  );
};

export default Settings;
