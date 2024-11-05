import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { IEnvironmentContext } from '../_layout';

const ProjectConfigAndSecrets = () => {
  const rootContext = useOutletContext<IEnvironmentContext>();
  return (
    <SidebarLayout
      navItems={[
        { label: 'Config', value: 'configs' },
        { label: 'Secrets', value: 'secrets' },
      ]}
      parentPath="/cs"
    >
      <Outlet
        context={{
          ...rootContext,
        }}
      />
    </SidebarLayout>
  );
};

export default ProjectConfigAndSecrets;
