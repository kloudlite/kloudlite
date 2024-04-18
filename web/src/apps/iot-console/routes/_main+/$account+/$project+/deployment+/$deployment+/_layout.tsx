import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { CommonTabs } from '~/iotconsole/components/common-navbar-tabs';
import { IDeployment } from '~/iotconsole/server/gql/queries/iot-deployment-queries';
import { IRemixCtx } from '~/root/lib/types/common';
import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import logger from '~/root/lib/client/helpers/log';
import { redirect } from '@remix-run/node';
import { IProjectContext } from '../../_layout';

export interface IDeploymentContext extends IProjectContext {
  deployment: IDeployment;
}

const Tabs = () => {
  const { account, project } = useParams();

  return (
    <CommonTabs
      backButton={{
        to: `/${account}/${project}/deployments`,
        label: 'Back to Deployment',
      }}
    />
  );
};
export const handle = () => {
  return {
    navbar: <Tabs />,
  };
};
const Deployment = () => {
  const rootContext = useOutletContext<IProjectContext>();
  const { deployment } = useLoaderData();
  return <Outlet context={{ ...rootContext, deployment }} />;
};

export default Deployment;

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { project, deployment, account } = ctx.params;

  try {
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).getIotDeployment({
      projectName: project,
      name: deployment,
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      deployment: data || {},
    };
  } catch (e) {
    logger.error(e);
    return redirect(`/${account}/projects`);
  }
};
