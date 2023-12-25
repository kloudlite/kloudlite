import { Outlet, useOutletContext } from '@remix-run/react';
import { IClusterContext } from '../_layout';
import SidebarLayout from '~/console/components/sidebar-layout';

const ContainerRegistry = () => {
  const rootContext = useOutletContext<IClusterContext>();
  return (
    <SidebarLayout
      navItems={[
        { label: 'Wireguard VPN', value: 'vpn' },
        { label: 'Domain', value: 'domain' },
      ]}
      parentPath="/network"
      // headerTitle="Network"
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default ContainerRegistry;
