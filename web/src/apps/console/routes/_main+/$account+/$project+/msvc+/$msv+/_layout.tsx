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
import { IProjectContext } from '~/console/routes/_main+/$account+/$project+/_layout';
import { IProjectMSv } from '~/console/server/gql/queries/project-managed-services-queries';

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

export const handle = () => {
  return {
    navbar: <ManagedServiceTabs />,
  };
};

export interface IProjectManagedServiceContext extends IProjectContext {
  managedService: IProjectMSv;
}

const MSOutlet = ({
  managedService: OProjectMSv,
}: {
  managedService: IProjectMSv;
}) => {
  const rootContext = useOutletContext<IProjectContext>();

  return <Outlet context={{ ...rootContext, managedService: OProjectMSv }} />;
};

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { msv, project } = ctx.params;
    try {
      const { data, errors } = await GQLServerHandler(
        ctx.request
      ).getProjectMSv({
        name: msv,
        projectName: project,
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
        managedService: {} as IProjectMSv,
        redirect: `../managed-services`,
      };
    }
  });
  return defer({ promise: await promise });
};

const ManagedService = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ managedService }) => {
        return <MSOutlet managedService={managedService} />;
      }}
    </LoadingComp>
  );
};

export default ManagedService;
