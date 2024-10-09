import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { useHandleFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import { IAccountContext } from '../_layout';

const Infra = () => {
  const rootContext = useOutletContext<IAccountContext>();
  const noLayout = useHandleFromMatches('noLayout', null);

  if (noLayout) {
    return <Outlet context={rootContext} />;
  }
  return (
    <SidebarLayout
      navItems={[
        { label: 'Attached Clusters', value: 'clusters' },
        // { label: 'Helm Repos', value: 'helm-repos' },
        // { label: 'Bring your own Kubernetes', value: 'byok-cluster' },
        { label: 'Wireguard Devices', value: 'vpn-devices' },
      ]}
      parentPath="/infra"
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default Infra;
