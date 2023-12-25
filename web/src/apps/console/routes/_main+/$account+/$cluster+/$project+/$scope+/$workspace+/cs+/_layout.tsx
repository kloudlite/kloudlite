import { Outlet, useOutletContext } from '@remix-run/react';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import { IWorkspaceContext } from '../_layout';
import SidebarLayout from '~/console/components/sidebar-layout';

const ProjectConfigAndSecrets = () => {
  const rootContext = useOutletContext<IWorkspaceContext>();
  const subNavAction = useSubNavData();
  return (
    <SidebarLayout
      headerActions={subNavAction.data}
      navItems={[
        { label: 'Config', value: 'configs' },
        { label: 'Secrets', value: 'secrets' },
      ]}
      parentPath="/cs"
      headerTitle="Settings"
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
