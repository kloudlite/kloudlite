import { Outlet } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';

const navItems = [
  { label: 'General', value: 'general' },
  { label: 'Compute', value: 'compute' },
  { label: 'Environment', value: 'environment' },
  { label: 'Network', value: 'network' },
];

const Settings = () => {
  return (
    <SidebarLayout
      navItems={navItems}
      parentPath="/settings"
      headerTitle="Settings"
    >
      <Outlet />
    </SidebarLayout>
  );
};

export default Settings;
