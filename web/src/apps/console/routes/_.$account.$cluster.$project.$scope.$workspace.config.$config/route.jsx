import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import Wrapper from '~/console/components/wrapper';
import { useParams, useLoaderData } from '@remix-run/react';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { getScopeAndProjectQuery } from '~/console/server/r-urils/common';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { defer } from '@remix-run/node';
import { useState } from 'react';
import { dummyData } from '~/console/dummy/data';
import Tools from './tools';
import Resources from './resources';
import Handle from './handle';

export const handle = () => {
  return {
    navbar: {},
  };
};

export const loader = async (ctx) => {
  // main promise
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    ensureClusterSet(ctx);

    const { config } = ctx.params;

    const { data, errors } = await GQLServerHandler(ctx.request).getConfig({
      name: config,
      ...getScopeAndProjectQuery(ctx),
    });

    if (errors) {
      throw errors[0];
    }
    return { config: data };
  });

  return defer({ promise });
};

const Config = () => {
  const [data, setData] = useState(dummyData.configs);
  const [showHandleConfig, setShowHandleConfig] = useState(null);
  const [modifiedData, setModifiedData] = useState(dummyData.configs);

  const { promise } = useLoaderData();
  const { account, cluster, project, scope, workspace } = useParams();

  const changesCount = () => {
    return modifiedData.filter(
      (md) =>
        md?.delete || md?.insert || (md?.newvalue && md.newvalue !== md.value)
    ).length;
  };

  const extractConfigs = () => {
    return modifiedData
      .filter((md) => !md?.delete)
      .map((md) => ({
        key: md.key,
        value: md?.newvalue ? md.newvalue : md.value,
      }));
  };

  return (
    <LoadingComp>
      {({ config }) => {
        // const { data } = config;
        return (
          <>
            <Wrapper
              header={{
                title: 'kloud-root-ca.crt',
                backurl: `/${account}/${cluster}/${project}/${scope}/${workspace}/cs/configs`,
                action: Object.keys(data).length > 0 && (
                  <div className="flex flex-row items-center gap-lg">
                    <Button
                      variant="outline"
                      content="Add new entry"
                      prefix={PlusFill}
                      onClick={() => setShowHandleConfig(true)}
                    />
                    {changesCount() > 0 && (
                      <Button
                        variant="basic"
                        content="Discard"
                        onClick={() => setModifiedData(data)}
                      />
                    )}
                    {changesCount() > 0 && (
                      <Button
                        variant="primary"
                        content={`Commit ${changesCount()} changes`}
                        onClick={() => console.log(extractConfigs())}
                      />
                    )}
                  </div>
                ),
              }}
              empty={{
                is: Object.keys(data).length === 0,
                title: 'This is where you’ll manage your projects.',
                content: (
                  <p>
                    You can create a new project and manage the listed project.
                  </p>
                ),
                action: {
                  content: 'Add new entry',
                  prefix: Plus,
                  onClick: () => setShowHandleConfig(true),
                },
              }}
            >
              <Tools />
              {/* <ResourceList>
              {Object.entries(data)?.map(([k, v]) => (
                <ResourceList.ResourceItem key={k} textValue={k}>
                  <Resources item={{ key: k, value: v }} />
                </ResourceList.ResourceItem>
              ))}
            </ResourceList> */}
              <Resources
                originalItems={data}
                modifiedItems={modifiedData}
                setModifiedData={setModifiedData}
              />
            </Wrapper>
            <Handle
              show={showHandleConfig}
              setShow={setShowHandleConfig}
              onSubmit={(val) => {
                setModifiedData((prev) => [{ ...val, insert: true }, ...prev]);
                setShowHandleConfig(false);
              }}
            />
          </>
        );
      }}
    </LoadingComp>
  );
};

export default Config;
