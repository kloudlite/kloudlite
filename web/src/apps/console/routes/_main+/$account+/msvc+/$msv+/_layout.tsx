import { defer } from '@remix-run/node';
import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import Breadcrum from '~/console/components/breadcrum';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import {
  BackingServices,
  ChevronRight,
  GearSix,
} from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IClusterMSv } from '~/console/server/gql/queries/cluster-managed-services-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseName } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { BreadcrumSlash, tabIconSize } from '~/console/utils/commons';
import logger from '~/lib/client/helpers/log';
import { IRemixCtx } from '~/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import { IAccountContext } from '../../_layout';

const ManagedServiceTabs = () => {
  const { account, msv } = useParams();
  const iconSize = tabIconSize;
  return (
    <CommonTabs
      baseurl={`/${account}/msvc/${msv}`}
      backButton={{
        to: `/${account}/managed-services`,
        label: 'Managed Services',
      }}
      tabs={[
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <BackingServices size={tabIconSize} />
              Managed resources
            </span>
          ),
          to: '/managed-resources',
          value: '/managed-resources',
        },
        {
          label: 'Logs & Metrics',
          to: '/logs-n-metrics',
          value: '/logs-n-metrics',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <GearSix size={iconSize} />
              Settings
            </span>
          ),
          to: '/settings/general',
          value: '/settings',
        },
      ]}
    />
  );
};

const LocalBreadcrum = ({ data }: { data: IClusterMSv }) => {
  const { displayName } = data;
  const { account } = useParams();
  return (
    <div className="flex flex-row items-center">
      <BreadcrumSlash />
      <span className="mx-md" />
      <Breadcrum.Button
        to={`/${account}/managed-services`}
        linkComponent={Link}
        content={
          <div className="flex flex-row gap-md items-center">
            Managed Services <ChevronRight size={14} />{' '}
          </div>
        }
      />
      <Breadcrum.Button
        to={`/${account}/msvc/${parseName(data)}/logs-n-metrics`}
        linkComponent={Link}
        content={<span>{displayName}</span>}
      />
    </div>
  );
};

export const handle = ({
  promise: { managedService, error },
}: {
  promise: any;
}) => {
  if (error) {
    return {};
  }

  return {
    navbar: <ManagedServiceTabs />,
    breadcrum: () => <LocalBreadcrum data={managedService} />,
  };
};

export interface IManagedServiceContext extends IAccountContext {
  managedService: IClusterMSv;
}

const MSOutlet = ({
  managedService: OClustMSv,
}: {
  managedService: IClusterMSv;
}) => {
  const rootContext = useOutletContext<IManagedServiceContext>();

  return <Outlet context={{ ...rootContext, managedService: OClustMSv }} />;
};

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { msv } = ctx.params;
    try {
      const { data, errors } = await GQLServerHandler(
        ctx.request
      ).getClusterMSv({
        name: msv,
      });
      if (errors) {
        throw errors[0];
      }

      return {
        managedService: data,
      };
    } catch (err) {
      logger.log(err);

      return {
        managedService: {} as IClusterMSv,
        redirect: `../managed-services`,
      };
    }
  });
  return defer({ promise: await promise });
};

const ManagedService = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp
      skeletonData={{
        managedService: fake.ConsoleListClusterMSvsQuery
          .infra_listClusterManagedServices as any,
      }}
      data={promise}
    >
      {({ managedService }) => {
        return <MSOutlet managedService={managedService} />;
      }}
    </LoadingComp>
  );
};

export default ManagedService;
