import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { IExternalAppContext } from '../_layout';

const Settings = () => {
  const rootContext = useOutletContext<IExternalAppContext>();
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
