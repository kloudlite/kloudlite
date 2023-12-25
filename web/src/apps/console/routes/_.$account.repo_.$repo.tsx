import { Link, Outlet, useOutletContext, useParams } from '@remix-run/react';
import { useSubNavData } from '~/root/lib/client/hooks/use-create-subnav-action';
import { ChevronRight } from '@jengaicons/react';
import SidebarLayout from '../components/sidebar-layout';
import { IProjectContext } from './_.$account.infra.$cluster.$project';
import { CommonTabs } from '../components/common-navbar-tabs';
import Breadcrum from '../components/breadcrum';
import { ExtractNodeType, parseName } from '../server/r-utils/common';
import { constants } from '../server/utils/constants';
import Wrapper from '../components/wrapper';

const NetworkBreadcrum = () => {
  const { repo, account } = useParams();
  return (
    <div className="flex flex-row items-center">
      <Breadcrum.Button
        to={`/${account}/packages`}
        LinkComponent={Link}
        content={
          <div className="flex flex-row gap-md items-center">
            <ChevronRight size={14} /> Packages <ChevronRight size={14} />{' '}
          </div>
        }
      />
      <Breadcrum.Button content={<span>{repo}</span>} />
    </div>
  );
};

export const handle = () => {
  return {
    navbar: constants.nan,
    breadcrum: () => <NetworkBreadcrum />,
  };
};

const Repo = () => {
  const rootContext = useOutletContext<IProjectContext>();
  return (
    // <SidebarLayout
    //   navItems={[
    //     { label: 'Images', value: 'images' },
    //     { label: 'Builds', value: 'builds' },
    //     { label: 'Build caches', value: 'buildcaches' },
    //   ]}
    //   parentPath={`/${repo}`}
    //   headerTitle={repo || ''}
    //   headerActions={subNavAction.data}
    // >
    //   <Outlet context={{ ...rootContext }} />
    // </SidebarLayout>
    <Wrapper
      header={{
        title: 'Images',
      }}
    >
      <Outlet context={{ ...rootContext }} />
    </Wrapper>
  );
};

export default Repo;
