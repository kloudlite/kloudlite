import { Outlet, useOutletContext } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import SidebarLayout from '../components/sidebar-layout';
import { IWorkspaceContext } from './_.$account.$cluster.$project.$scope.$workspace/route';

const ProjectConfigAndSecrets = () => {
  const rootContext = useOutletContext<IWorkspaceContext>();
  const subNavAction = useSubNavData();
  return (
    <SidebarLayout
      headerActions={
        subNavAction.data &&
        subNavAction.data.show && (
          <Button
            variant="primary"
            content={subNavAction.data.content}
            onClick={subNavAction.data.action}
          />
        )
      }
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
