import { Outlet, useMatches } from '@remix-run/react';
import { SubHeader } from '~/components/organisms/sub-header';
import * as ActionList from '~/components/atoms/action-list';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';

const sidenav = [
  { label: 'Config', href: 'config' },
  { label: 'Secrets', href: 'secrets' },
];
const ProjectConfigAndSecrets = () => {
  const { activePath } = useActivePath({ parent: '/config-and-secrets' });
  const [subNavAction, setSubNavAction] = useState(null);
  const ActionMatch = useMatches();

  let ReceivedButton = ActionMatch.reverse().find(
    (m) => m.handle?.subheaderAction
  )?.handle?.subheaderAction;
  ReceivedButton = ReceivedButton();

  return (
    <>
      <SubHeader
        title="Config & Secrets"
        actions={
          <Button
            {...ReceivedButton.props}
            onClick={() => {
              if (subNavAction) {
                subNavAction.action();
              }
            }}
          />
        }
      />
      <div className="flex flex-row gap-10xl">
        <div className="w-[180px]">
          <ActionList.ActionRoot value={activePath}>
            {sidenav.map((sn) => (
              <ActionList.ActionButton
                value={`/${sn.href}`}
                href={sn.href}
                key={sn.href}
              >
                {sn.label}
              </ActionList.ActionButton>
            ))}
          </ActionList.ActionRoot>
        </div>
        <div className="flex-1 flex flex-col gap-6xl">
          <Outlet context={[subNavAction, setSubNavAction]} />
        </div>
      </div>
    </>
  );
};

export default ProjectConfigAndSecrets;
