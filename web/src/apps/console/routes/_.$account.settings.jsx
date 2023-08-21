import { Outlet } from '@remix-run/react';
import SidebarLayout from '../components/sidebar-layout';

const Settings = () => {
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        { label: 'User management', value: 'user-management' },
      ]}
      parentPath="/settings"
      headerTitle="Settings"
    >
      <Outlet />
    </SidebarLayout>
  );
};

export default Settings;
