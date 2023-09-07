import { Outlet, useMatches, useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';

import { AnimatePresence } from 'framer-motion';
import SidebarLayout from '../components/sidebar-layout';
import { IWorkspaceContext } from './_.$account.$cluster.$project.$scope.$workspace/route';

const ProjectConfigAndSecrets = () => {
  const [subNavAction, setSubNavAction] = useState<any | null>(null);
  const rootContext = useOutletContext<IWorkspaceContext>();
  const ActionMatch = useMatches();

  let ReceivedButton = ActionMatch.reverse().find(
    (m) => m?.handle?.subheaderAction
  )?.handle?.subheaderAction;
  ReceivedButton = ReceivedButton();
  return (
    <SidebarLayout
      headerActions={
        <Button
          {...ReceivedButton.props}
          onClick={() => {
            if (subNavAction) {
              subNavAction.action();
            }
          }}
        />
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
            subNavAction,
            setSubNavAction,

            ...rootContext,
          }}
        />
      </AnimatePresence>
    </SidebarLayout>
  );
};

export default ProjectConfigAndSecrets;
