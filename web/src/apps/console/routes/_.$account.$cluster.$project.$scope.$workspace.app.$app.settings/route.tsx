import { Outlet, useOutletContext } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import SidebarLayout from '~/console/components/sidebar-layout';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import { IAppContext } from '../_.$account.$cluster.$project.$scope.$workspace.app.$app/route';

const navItems = [
  { label: 'General', value: 'general' },
  { label: 'Compute', value: 'compute' },
  { label: 'Environment', value: 'environment' },
  { label: 'Network', value: 'network' },
  { label: 'Advance', value: 'advance' },
];

const Settings = () => {
  const rootContext = useOutletContext<IAppContext>();
  const { data: subNavAction } = useSubNavData();
  return (
    <SidebarLayout
      navItems={navItems}
      parentPath="/settings"
      headerTitle="Settings"
      headerActions={
        subNavAction &&
        subNavAction.show && (
          <div className="flex flex-row items-center gap-lg">
            <Button
              variant="basic"
              content="Discard"
              onClick={subNavAction.subAction}
            />
            <Button
              variant="primary"
              content={subNavAction.content}
              onClick={subNavAction.action}
            />
          </div>
        )
      }
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default Settings;
