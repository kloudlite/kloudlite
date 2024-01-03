import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';

const ManagedServicesLayout = () => {
  const rootContext = useOutletContext();
  //   const noLayout = useHandleFromMatches('noLayout', null);

  //   if (noLayout) {
  //     return <Outlet context={rootContext} />;
  //   }
  return (
    <SidebarLayout
      navItems={[
        { label: 'KlOperator Services', value: 'kl-operator-services' },
        { label: 'Helm Charts', value: 'helm-chart' },
      ]}
      parentPath="/managedservices"
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default ManagedServicesLayout;
