import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '../components/sidebar-layout';
import { IProjectContext } from './_.$account.$cluster.$project';

const Settings = () => {
  const rootContext = useOutletContext<IProjectContext>();
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        { label: 'Access management', value: 'access-management' },
      ]}
      parentPath="/settings"
      headerTitle="Settings"
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default Settings;
