import { Outlet, useOutletContext } from '@remix-run/react';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import SidebarLayout from '../components/sidebar-layout';
import { IWorkspaceContext } from './_.$account.$cluster.$project.$scope.$workspace/route';

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
