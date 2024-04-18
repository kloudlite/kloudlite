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
import { BreadcrumSlash, tabIconSize } from '~/iotconsole/utils/commons';
import { GearSix, VirtualMachine } from '~/iotconsole/components/icons';
import Breadcrum from '~/iotconsole/components/breadcrum';
import { Truncate } from '~/root/lib/utils/common';
import { IProjectContext } from '../../_layout';

export interface IDeploymentContext extends IProjectContext {
  deployment: IDeployment;
}

const iconSize = tabIconSize;
const tabs = [
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <VirtualMachine size={iconSize} />
        Devices
      </span>
    ),
    to: '/devices',
    value: '/devices',
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
];

const LocalBreadcrum = ({ data }: { data: IDeployment }) => {
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

const Tabs = () => {
  const { account, project, deployment } = useParams();

  return (
    <CommonTabs
      baseurl={`/${account}/${project}/deployment/${deployment}`}
      tabs={tabs}
    />
  );

  // return (
  //   <CommonTabs
  //     backButton={{
  //       to: `/${account}/${project}/deployments`,
  //       label: 'Back to Deployment',
  //     }}
  //   />
  // );
};
export const handle = ({ deployment }: { deployment: any }) => {
  return {
    navbar: <Tabs />,
    breadcrum: () => <LocalBreadcrum data={deployment} />,
  };
};

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

const Deployment = () => {
  const rootContext = useOutletContext<IProjectContext>();
  const { deployment } = useLoaderData();
  return <Outlet context={{ ...rootContext, deployment }} />;
};

export default Deployment;
