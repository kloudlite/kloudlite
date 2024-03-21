import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { IEnvironmentContext } from '../_layout';

const WorkspaceSettings = () => {
  const rootContext = useOutletContext<IEnvironmentContext>();

  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        // { label: 'Image Pull Secrets', value: 'image-pull-secrets' },
      ]}
      parentPath="/settings"
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default WorkspaceSettings;
