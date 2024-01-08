import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';

const ClusterSettings = () => {
  const rootContext = useOutletContext();
  const subNavAction = useSubNavData();
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        { label: 'Domains', value: 'domain' },
      ]}
      parentPath="/settings"
      headerActions={subNavAction.data}
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default ClusterSettings;
