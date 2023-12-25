import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '../components/sidebar-layout';
import { IClusterContext } from './_.$account.infra.$cluster';

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
