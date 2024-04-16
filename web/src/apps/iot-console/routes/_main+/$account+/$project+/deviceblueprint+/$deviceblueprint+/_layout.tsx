import { Outlet, useOutletContext, useParams } from '@remix-run/react';
import { CommonTabs } from '~/iotconsole/components/common-navbar-tabs';
import { tabIconSize } from '~/iotconsole/utils/commons';
import { VirtualMachine } from '~/iotconsole/components/icons';

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
  // {
  //   label: (
  //     <span className="flex flex-row items-center gap-lg">
  //       <GearSix size={iconSize} />
  //       Settings
  //     </span>
  //   ),
  //   to: '/settings/general',
  //   value: '/settings',
  // },
];

const Tabs = () => {
  const { account, project, deviceblueprint } = useParams();

  return (
    <CommonTabs
      baseurl={`/${account}/${project}/deviceblueprint/${deviceblueprint}`}
      tabs={tabs}
    />
  );

  // return (
  //   <CommonTabs
  //     backButton={{
  //       to: `/${account}/${project}/deviceblueprints`,
  //       label: 'Back to Device Blueprint',
  //     }}
  //   />
  // );
};
export const handle = () => {
  return {
    navbar: <Tabs />,
  };
};
const DeviceBlueprint = () => {
  const rootContext = useOutletContext();
  return <Outlet context={rootContext} />;
};

export default DeviceBlueprint;
