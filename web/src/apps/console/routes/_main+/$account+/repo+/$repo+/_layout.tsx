import { Link, Outlet, useOutletContext, useParams } from '@remix-run/react';
import {
  ChevronRight,
  GearSix,
  GitMerge,
  NoOps,
  Nodeless,
} from '@jengaicons/react';
import Breadcrum from '~/console/components/breadcrum';
import { CommonTabs } from '~/console/components/common-navbar-tabs';

import { IProjectContext } from '../../$cluster+/$project+/_layout';

const LocalBreadcrum = () => {
  const { repo, account } = useParams();
  return (
    <div className="flex flex-row items-center">
      <Breadcrum.Button
        to={`/${account}/packages`}
        LinkComponent={Link}
        content={
          <div className="flex flex-row gap-md items-center">
            <ChevronRight size={14} />{' '}
            <div className="flex flex-row items-center gap-lg">
              Container Repos
            </div>
            <ChevronRight size={14} />{' '}
          </div>
        }
      />
      <Breadcrum.Button
        to={`/${account}/repo/${repo}`}
        LinkComponent={Link}
        content={<span>{repo}</span>}
      />
    </div>
  );
};

const Tabs = () => {
  const { repo, account } = useParams();
  const iconSize = 16;
  return (
    <CommonTabs
      baseurl={`/${account}/repo/${repo}`}
      tabs={[
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <Nodeless size={iconSize} />
              Images
            </span>
          ),
          value: '/images',
          to: '/images',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <GitMerge size={iconSize} />
              Build Integrations
            </span>
          ),
          value: '/builds',
          to: '/builds',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <NoOps size={iconSize} />
              Build Runs
            </span>
          ),
          value: '/buildruns',
          to: '/buildruns',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <GearSix size={iconSize} />
              Settings
            </span>
          ),
          value: '/settings',
          to: '/settings',
        },
      ]}
    />
  );
};
export const handle = () => {
  return {
    navbar: <Tabs />,
    breadcrum: () => <LocalBreadcrum />,
  };
};

const Repo = () => {
  const rootContext = useOutletContext<IProjectContext>();
  return <Outlet context={{ ...rootContext }} />;
};

export default Repo;
