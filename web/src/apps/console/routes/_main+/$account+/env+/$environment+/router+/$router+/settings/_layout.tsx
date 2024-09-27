import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { IEnvironmentContext } from '../../../_layout';

const Settings = () => {
  const rootContext = useOutletContext<IEnvironmentContext>();
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
