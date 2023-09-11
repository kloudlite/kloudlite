import { useParams, Outlet, useLocation } from '@remix-run/react';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { IRemixCtx } from '~/root/lib/types/common';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import useBasepath from '~/root/lib/client/hooks/use-basepath';

const ProjectTabs = () => {
  const { account, cluster, project, workspace } = useParams();
  // const { path, getPrevious } = useBasepath();
  // console.log('previous', getPrevious());

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

const App = () => {
  return <Outlet />;
};

export default App;
