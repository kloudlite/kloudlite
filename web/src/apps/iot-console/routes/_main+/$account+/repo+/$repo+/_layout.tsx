import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import {
  ChevronRight,
  GearSix,
  GitMerge,
  NoOps,
  Nodeless,
} from '~/iotconsole/components/icons';
import Breadcrum from '~/iotconsole/components/breadcrum';
import { CommonTabs } from '~/iotconsole/components/common-navbar-tabs';

import { IRemixCtx, LoaderResult } from '~/root/lib/types/common';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import logger from '~/root/lib/client/helpers/log';
import { IPackageContext } from '~/iotconsole/routes/_main+/$account+/packages+/_layout';

const LocalBreadcrum = () => {
  const { account, repo } = useParams();
  return (
    <div className="flex flex-row items-center">
      <Breadcrum.Button
        to={`/${account}/packages`}
        linkComponent={Link}
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
        linkComponent={Link}
        content={<span>{atob(repo || '')}</span>}
      />
    </div>
  );
};

const Tabs = () => {
  const { account } = useParams();

  const { repo } = useParams();
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
              Buildruns
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

export const loader = async (ctx: IRemixCtx) => {
  const { repo } = ctx.params;

  const repoName = atob(repo || '');

  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getLogins({});

    if (errors) {
      throw errors[0];
    }

    const { data: e, errors: dErrors } = await GQLServerHandler(
      ctx.request
    ).loginUrls({});

    if (dErrors) {
      throw dErrors[0];
    }

    return {
      loginUrls: e,
      logins: data,
      repoName,
    };
  } catch (err) {
    logger.error(err);
  }

  const k: any = {};

  return {
    logins: k,
    loginUrls: k,
    repoName,
  };
};

export const handle = () => {
  return {
    navbar: <Tabs />,
    breadcrum: () => <LocalBreadcrum />,
  };
};

export interface IRepoContext extends IPackageContext {
  logins: LoaderResult<typeof loader>['logins'];
  loginUrls: LoaderResult<typeof loader>['loginUrls'];
  repoName: LoaderResult<typeof loader>['repoName'];
}

const Repo = () => {
  const rootContext = useOutletContext<IPackageContext>();
  const ctx = useLoaderData<typeof loader>();
  return <Outlet context={{ ...rootContext, ...ctx }} />;
};

export default Repo;
