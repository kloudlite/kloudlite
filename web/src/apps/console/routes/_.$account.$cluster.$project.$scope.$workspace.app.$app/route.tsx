import { useParams, Outlet, useOutletContext } from '@remix-run/react';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { IRemixCtx } from '~/root/lib/types/common';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import useBasepath from '~/root/lib/client/hooks/use-basepath';
import { IApp } from '~/console/server/gql/queries/app-queries';
import { SubNavDataProvider } from '~/root/lib/client/hooks/use-create-subnav-action';
import { IWorkspaceContext } from '../_.$account.$cluster.$project.$scope.$workspace/route';

const ProjectTabs = () => {
  const { account, cluster, project, workspace } = useParams();

  return (
    <CommonTabs
      baseurl=""
      backButton={{
        to: '',
        label: 'Apps',
      }}
      tabs={[
        {
          label: 'Overview',
          to: '/overview',
          value: '/overview',
        },
        {
          label: 'Logs',
          to: '/logs',
          value: '/logs',
        },

        {
          label: 'Settings',
          to: '/settings/general',
          value: '/settings',
        },
      ]}
    />
  );
};

export const handle = () => {
  return {
    navbar: <ProjectTabs />,
  };
};

export interface IAppContext extends IWorkspaceContext {
  app: IApp;
}
const App = () => {
  const rootContext = useOutletContext<IWorkspaceContext>();
  return (
    <SubNavDataProvider>
      <Outlet context={{ ...rootContext }} />
    </SubNavDataProvider>
  );
};

export default App;
