import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '../components/sidebar-layout';
import { IClusterContext } from './_.$account.$cluster';

const ContainerRegistry = () => {
  const rootContext = useOutletContext<IClusterContext>();
  return (
    <SidebarLayout
      navItems={[
        { label: 'Domain', value: 'domain' },
        { label: 'VPN', value: 'vpn' },
      ]}
      parentPath="/network"
      headerTitle="Network"
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default ContainerRegistry;
