import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';

const Settings = () => {
  const rootContext = useOutletContext<any>();
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        { label: 'Build caches', value: 'buildcaches' },
      ]}
      parentPath="/settings"
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default Settings;
