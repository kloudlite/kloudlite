import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import { useHandleFromMatches } from '~/root/lib/client/hooks/use-custom-matches';

const Infra = () => {
  const rootContext = useOutletContext();
  const subNavAction = useSubNavData();
  const noLayout = useHandleFromMatches('noLayout', null);

  if (noLayout) {
    return <Outlet context={rootContext} />;
  }
  return (
    <SidebarLayout
      navItems={[
        { label: 'k8s Clusters', value: 'clusters' },
        { label: 'VMs', value: 'vms' },
      ]}
      parentPath="/infra"
      headerActions={subNavAction.data}
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default Infra;
