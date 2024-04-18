import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/iotconsole/components/sidebar-layout';

const Settings = () => {
  const rootContext = useOutletContext<any>();
  return (
    <SidebarLayout
      navItems={[{ label: 'General', value: 'general' }]}
      parentPath="/settings"
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default Settings;
