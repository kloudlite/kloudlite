import { Outlet, useOutletContext } from '@remix-run/react';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import { IWorkspaceContext } from '../_layout';
import SidebarLayout from '~/console/components/sidebar-layout';

const WorkspaceSettings = () => {
  const rootContext = useOutletContext<IWorkspaceContext>();
  const subNavAction = useSubNavData();

  return (
    <SidebarLayout
      navItems={[{ label: 'General', value: 'general' }]}
      parentPath="/settings"
      headerTitle="Settings"
      headerActions={subNavAction.data}
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default WorkspaceSettings;
