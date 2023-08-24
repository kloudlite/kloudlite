import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '../components/sidebar-layout';

const Settings = () => {
  const rootContext = useOutletContext();
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        { label: 'User management', value: 'user-management' },
      ]}
      parentPath="/settings"
      headerTitle="Settings"
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default Settings;
