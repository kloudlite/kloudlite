import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/iotconsole/components/sidebar-layout';
import { IProjectContext } from '../_layout';

const Settings = () => {
  const rootContext = useOutletContext<IProjectContext>();
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
