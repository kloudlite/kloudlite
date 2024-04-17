import { Plus } from '~/iotconsole/components/icons';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import {
  LoadingComp,
  pWrapper,
} from '~/iotconsole/components/loading-component';
import Wrapper from '~/iotconsole/components/wrapper';
import { parseNodes } from '~/iotconsole/server/r-utils/common';
import { getPagination, getSearch } from '~/iotconsole/server/utils/common';
import { IRemixCtx } from '~/root/lib/types/common';

import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import Tools from './tools';
import DeviceBlueprintResource from './device-blueprint-resource';
import HandleDeviceBlueprint from './handle-deviceblueprint';

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  const { project } = ctx.params;
  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listIotDeviceBlueprints({
      projectName: project,
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });

    if (errors) {
      throw errors[0];
    }

    return {
      deviceBlueprintData: data || {},
    };
  });

  return defer({ promise });
};

const Workspaces = () => {
  const [visible, setVisible] = useState(false);

  const { promise } = useLoaderData<typeof loader>();
  return (
    <>
      <LoadingComp
        data={promise}
        // skeletonData={{
        //   environmentData: fake.ConsoleListEnvironmentsQuery
        //     .core_listEnvironments as any,
        // }}
      >
        {({ deviceBlueprintData }) => {
          console.log(deviceBlueprintData);

          const deviceBlueprints = parseNodes(deviceBlueprintData);

          if (!deviceBlueprints) {
            return null;
          }

          return (
            <Wrapper
              header={{
                title: 'Device Blueprints',
                action: deviceBlueprints.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create Device Blueprint"
                    prefix={<Plus />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                is: deviceBlueprints?.length === 0,
                title: 'This is where you’ll manage your device blueprint.',
                content: (
                  <p>
                    You can create a new device blueprint and manage the listed
                    device blueprints.
                  </p>
                ),
                action: {
                  content: 'Create new device blueprint',
                  prefix: <Plus />,
                  LinkComponent: Link,
                  onClick: () => {
                    setVisible(true);
                  },
                },
              }}
              tools={<Tools />}
            >
              <DeviceBlueprintResource items={deviceBlueprints || []} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleDeviceBlueprint
        {...{
          visible,
          setVisible,
          isUpdate: false,
        }}
      />
    </>
  );
};
export default Workspaces;
