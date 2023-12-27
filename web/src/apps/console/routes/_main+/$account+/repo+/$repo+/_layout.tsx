import { Link, Outlet, useOutletContext, useParams } from '@remix-run/react';
import { ChevronRight } from '@jengaicons/react';
import Breadcrum from '~/console/components/breadcrum';
import { constants } from '~/console/server/utils/constants';
import Wrapper from '~/console/components/wrapper';
import { IProjectContext } from '../../$cluster+/$project+/_layout';

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
