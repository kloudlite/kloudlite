import { Outlet, useOutletContext, useParams } from '@remix-run/react';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import SidebarLayout from '../components/sidebar-layout';
import { IProjectContext } from './_.$account.$cluster.$project';

const Repo = () => {
  const rootContext = useOutletContext<IProjectContext>();
  const subNavAction = useSubNavData();

  const { repo } = useParams();

  return (
    <SidebarLayout
      navItems={[
        { label: 'Images', value: 'images' },
        { label: 'Builds', value: 'builds' },
        { label: 'Build caches', value: 'buildcaches' },
      ]}
      parentPath={`/${repo}`}
      headerTitle={repo || ''}
      headerActions={subNavAction.data}
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default Repo;
