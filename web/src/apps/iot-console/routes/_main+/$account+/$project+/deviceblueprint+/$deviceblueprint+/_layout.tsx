import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { CommonTabs } from '~/iotconsole/components/common-navbar-tabs';
import { BreadcrumSlash, tabIconSize } from '~/iotconsole/utils/commons';
import { GearSix, VirtualMachine } from '~/iotconsole/components/icons';
import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import logger from '~/root/lib/client/helpers/log';
import { redirect } from 'react-router-dom';
import { IRemixCtx } from '~/root/lib/types/common';
import { IDeviceBlueprint } from '~/iotconsole/server/gql/queries/iot-device-blueprint-queries';
import Breadcrum from '~/iotconsole/components/breadcrum';
import { Truncate } from '~/root/lib/utils/common';
import { IProjectContext } from '../../_layout';

export interface IDeviceBlueprintContext extends IProjectContext {
  deviceblueprint: IDeviceBlueprint;
}

const iconSize = tabIconSize;
const tabs = [
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <VirtualMachine size={iconSize} />
        Apps
      </span>
    ),
    to: '/apps',
    value: '/apps',
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

const LocalBreadcrum = ({ data }: { data: IDeviceBlueprint }) => {
  const { displayName, accountName, name } = data;
  return (
    <div className="flex flex-row items-center">
      <BreadcrumSlash />
      <Breadcrum.Button
        content={<Truncate length={15}>{displayName || ''}</Truncate>}
        to={`/${accountName}/${projectName}/deviceblueprint/${name}/apps`}
      />
    </div>
  );
};

const Tabs = () => {
  const { account, deviceblueprint } = useParams();

  return (
    <CommonTabs
      baseurl={`/${account}/deviceblueprint/${deviceblueprint}`}
      tabs={tabs}
    />
  );

  // return (
  //   <CommonTabs
  //     backButton={{
  //       to: `/${account}/deviceblueprints`,
  //       label: 'Back to Device Blueprint',
  //     }}
  //   />
  // );
};
export const handle = ({ deviceblueprint }: { deviceblueprint: any }) => {
  return {
    navbar: <Tabs />,
    breadcrum: () => <LocalBreadcrum data={deviceblueprint} />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { deviceblueprint, account } = ctx.params;

  try {
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).getIotDeviceBlueprint({
      name: deviceblueprint,
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      deviceblueprint: data || {},
    };
  } catch (e) {
    logger.error(e);
    return redirect(`/${account}/environments`);
  }
};

const DeviceBlueprint = () => {
  const rootContext = useOutletContext<IProjectContext>();
  const { deviceblueprint } = useLoaderData();
  return <Outlet context={{ ...rootContext, deviceblueprint }} />;
};

export default DeviceBlueprint;
