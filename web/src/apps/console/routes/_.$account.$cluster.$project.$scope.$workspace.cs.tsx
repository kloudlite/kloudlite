import { Outlet, useOutletContext } from '@remix-run/react';
import { AnimatePresence } from 'framer-motion';
import { Button } from '~/components/atoms/button';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import SidebarLayout from '../components/sidebar-layout';
import { IWorkspaceContext } from './_.$account.$cluster.$project.$scope.$workspace/route';

const ProjectConfigAndSecrets = () => {
  const rootContext = useOutletContext<IWorkspaceContext>();
  const { data: subNavAction } = useSubNavData();
  return (
    <SidebarLayout
      headerActions={
        subNavAction && (
          <Button
            variant="primary"
            content={subNavAction.content}
            onClick={subNavAction.action}
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
      <AnimatePresence mode="wait">
        <Outlet
          context={{
            ...rootContext,
          }}
        />
      </AnimatePresence>
    </SidebarLayout>
  );
};

export default ProjectConfigAndSecrets;
