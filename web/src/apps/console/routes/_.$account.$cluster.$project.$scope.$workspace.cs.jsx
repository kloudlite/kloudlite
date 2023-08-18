import { Outlet, useMatches } from '@remix-run/react';
import { SubHeader } from '~/components/organisms/sub-header';
import ActionList from '~/components/atoms/action-list';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';

const sidenav = [
  { label: 'Config', href: 'configs' },
  { label: 'Secrets', href: 'secrets' },
];

const ProjectConfigAndSecrets = () => {
  const { activePath } = useActivePath({ parent: '/cs' });
  const [subNavAction, setSubNavAction] = useState(null);
  const ActionMatch = useMatches();

  let ReceivedButton = ActionMatch.reverse().find(
    (m) => m.handle?.subheaderAction
  )?.handle?.subheaderAction;
  ReceivedButton = ReceivedButton();

  console.log(activePath);
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
        <div className="w-[180px] pt-3xl">
          <ActionList.Root value={activePath}>
            {sidenav.map((sn) => (
              <ActionList.Button
                value={`/${sn.href}`}
                href={sn.href}
                key={sn.href}
              >
                {sn.label}
              </ActionList.Button>
            ))}
          </ActionList.Root>
        </div>
        <div className="flex-1 flex flex-col gap-6xl">
          <Outlet context={{ subNavAction, setSubNavAction }} />
        </div>
      </div>
    </>
  );
};

export default ProjectConfigAndSecrets;
