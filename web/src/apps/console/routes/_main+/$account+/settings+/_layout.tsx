import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';

const Settings = () => {
  const rootContext = useOutletContext();
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        { label: 'User management', value: 'user-management' },
        { label: 'Cloud providers', value: 'cloud-providers' },
        { label: 'VPN', value: 'vpn' },
      ]}
      parentPath="/settings"
      // headerTitle="Settings"
      // headerActions={subNavAction.data}
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default Settings;
