import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { IManagedServiceContext } from '../_layout';

const Settings = () => {
  const rootContext = useOutletContext<IManagedServiceContext>();
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        // { label: 'Access management', value: 'access-management' },
      ]}
      parentPath="/settings"
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default Settings;
