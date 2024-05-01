import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { IRemixCtx } from '~/lib/types/common';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import logger from '~/lib/client/helpers/log';
import { defer } from '@remix-run/node';
import Breadcrum from '~/console/components/breadcrum';
import { BreadcrumSlash } from '~/console/utils/commons';
import { Truncate } from '~/root/lib/utils/common';
import { IClusterMSv } from '~/console/server/gql/queries/cluster-managed-services-queries';
import fake from '~/root/fake-data-generator/fake';
import { IClusterContext } from '../../_layout';

const ManagedServiceTabs = () => {
  const { account, project, msv } = useParams();
  return (
    <CommonTabs
      baseurl={`/${account}/${project}/msvc/${msv}`}
      backButton={{
        to: `/${account}/${project}/managed-services`,
        label: 'Managed Services',
      }}
      tabs={[
        {
          label: 'Logs & Metrics',
          to: '/logs-n-metrics',
          value: '/logs-n-metrics',
        },
      ]}
    />
  );
};

const LocalBreadcrum = ({ data }: { data: IClusterMSv }) => {
  const { displayName } = data;
  return (
    <div className="flex flex-row items-center">
      <BreadcrumSlash />
      <Breadcrum.Button
        content={<Truncate length={15}>{displayName || ''}</Truncate>}
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

export interface IManagedServiceContext extends IClusterContext {
  managedService: IClusterMSv;
}

const MSOutlet = ({
  managedService: OProjectMSv,
}: {
  managedService: IClusterMSv;
}) => {
  const rootContext = useOutletContext<IManagedServiceContext>();

  return <Outlet context={{ ...rootContext, managedService: OProjectMSv }} />;
};

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { msv } = ctx.params;
    try {
      const { data, errors } = await GQLServerHandler(
        ctx.request
      ).getClusterMSv({
        clusterName: ctx.params.cluster,
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
  return defer({ promise });
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
