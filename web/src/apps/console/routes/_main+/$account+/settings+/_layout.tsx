import { Outlet, useLoaderData, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { useHandleFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
import { IExtRemixCtx, LoaderResult } from '~/root/lib/types/common';
import { IAccountContext } from '../_layout';

const Settings = () => {
  // const rootContext = useOutletContext();
  const rootContext = useOutletContext<IAccountContext>();
  const noLayout = useHandleFromMatches('noLayout', null);

  const { teamMembers, currentUser } = useLoaderData();

  if (noLayout) {
    return <Outlet context={rootContext} />;
  }
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        { label: 'User management', value: 'user-management' },
        // { label: 'Cloud providers', value: 'cloud-providers' },
        { label: 'Image pull secrets', value: 'image-pull-secrets' },
        { label: 'Image Discovery', value: 'images' },
        // { label: 'VPN', value: 'vpn' },
      ]}
      parentPath="/settings"
      // headerTitle="Settings"
      // headerActions={subNavAction.data}
    >
      <Outlet
        context={{
          ...rootContext,
          teamMembers,
          currentUser,
        }}
      />
    </SidebarLayout>
  );
};

export const loader = async (ctx: IExtRemixCtx) => {
  const { account } = ctx.params;
  ensureAccountSet(ctx);
  const { data, errors } = await GQLServerHandler(
    ctx.request
  ).listMembershipsForAccount({
    accountName: account,
  });
  if (errors) {
    throw errors[0];
  }

  const { data: currentUser, errors: cErrors } = await GQLServerHandler(
    ctx.request
  ).whoAmI({});

  if (cErrors) {
    throw cErrors[0];
  }

  return {
    teamMembers: data,
    currentUser,
  };
};

export interface ISettingsContext extends IAccountContext {
  teamMembers: LoaderResult<typeof loader>['teamMembers'];
  currentUser: LoaderResult<typeof loader>['currentUser'];
}

export default Settings;
