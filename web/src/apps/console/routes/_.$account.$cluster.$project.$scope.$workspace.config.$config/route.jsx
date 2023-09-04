import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import Wrapper from '~/console/components/wrapper';
import { useParams, useLoaderData, useOutletContext } from '@remix-run/react';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseName } from '~/console/server/r-urils/common';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { defer } from '@remix-run/node';
import { useEffect, useState } from 'react';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { constants } from '~/console/server/utils/constants';
import { getScopeAndProjectQuery } from '~/console/server/utils/common';
import Tools from './tools';
import Resources from './resources';
import Handle, { updateConfig } from './handle';

export const handle = () => {
  return {
    navbar: constants.nan,
  };
};

// @ts-ignore
const DataSetter = ({ set = (/** @type {any} */ _) => _, value }) => {
  useEffect(() => {
    console.log(value);
    set(value);
  }, [value]);
  return null;
};
export const loader = async (
  /** @type {{ params: { config: any; }; request: { headers: any; cookies: any; }; }} */ ctx
) => {
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
  const [showHandleConfig, setShowHandleConfig] = useState(null);
  const [originalItems, setOriginalItems] = useState({});
  const [modifiedItems, setModifiedItems] = useState({});
  const [configUpdating, setConfigUpdating] = useState(false);
  const { promise } = useLoaderData();
  const { account, cluster, project, scope, workspace } = useParams();

  const api = useAPIClient();
  const context = useOutletContext();
  const reload = useReload();

  useEffect(() => {
    setModifiedItems(
      Object.entries(originalItems).reduce((acc, [key, value]) => {
        return {
          ...acc,
          [key]: {
            value,
            delete: false,
            edit: false,
            insert: false,
            newvalue: null,
          },
        };
      }, {})
    );
  }, [originalItems]);

  // @ts-ignore
  const addItem = ({ key, val }) => {
    setModifiedItems((prev) => ({
      [key]: {
        value: val,
        insert: true,
        delete: false,
        edit: false,
      },
      ...prev,
    }));
  };

  // @ts-ignore
  const deleteItem = ({ key, value }) => {
    // @ts-ignore
    if (originalItems[key]) {
      setModifiedItems((prev) => ({
        ...prev,
        [key]: { ...value, delete: true },
      }));
    } else {
      const mItems = { ...modifiedItems };
      // @ts-ignore
      delete mItems[key];
      setModifiedItems(mItems);
    }
  };

  // @ts-ignore
  const editItem = ({ key, value }, val) => {
    // @ts-ignore
    if (modifiedItems[key].insert) {
      setModifiedItems((prev) => ({
        ...prev,
        [key]: { ...value, value: val },
      }));
    } else {
      setModifiedItems((prev) => ({
        ...prev,
        [key]: { ...value, newvalue: val },
      }));
    }
  };

  // @ts-ignore
  const restoreItem = ({ key }) => {
    setModifiedItems((prev) => ({
      ...prev,
      [key]: {
        // @ts-ignore
        value: originalItems[key],
        delete: false,
        insert: false,
      },
    }));
  };

  const changesCount = () => {
    // return modifiedItems.filter(
    //   (md) =>
    //     md?.delete || md?.insert || (md?.newvalue && md.newvalue !== md.value)
    // ).length;
    return Object.values(modifiedItems).filter(
      (mi) =>
        mi.delete ||
        mi.insert ||
        (mi.newvalue != null && mi.newvalue !== mi.value)
    ).length;
  };

  return (
    <LoadingComp data={promise}>
      {({ config }) => {
        const { data: d } = config;
        return (
          <>
            <DataSetter set={setOriginalItems} value={d} />
            <Wrapper
              header={{
                title: parseName(config),
                backurl: `/${account}/${cluster}/${project}/${scope}/${workspace}/cs/configs`,
                action: Object.keys(modifiedItems).length > 0 && (
                  <div className="flex flex-row items-center gap-lg">
                    <Button
                      variant="outline"
                      content="Add new entry"
                      prefix={<PlusFill />}
                      onClick={() =>
                        // @ts-ignore
                        setShowHandleConfig({ data: modifiedItems })
                      }
                    />
                    {changesCount() > 0 && (
                      <Button variant="basic" content="Discard" />
                    )}
                    {changesCount() > 0 && (
                      <Button
                        variant="primary"
                        content={`Commit ${changesCount()} changes`}
                        loading={configUpdating}
                        onClick={async () => {
                          setConfigUpdating(true);
                          const k = Object.entries(modifiedItems).reduce(
                            (acc, [key, val]) => {
                              if (val.delete) {
                                return { ...acc };
                              }
                              return {
                                ...acc,
                                [key]: val.newvalue ? val.newvalue : val.value,
                              };
                            },
                            {}
                          );

                          await updateConfig({
                            api,
                            context,
                            config,
                            data: k,
                            reload,
                          });
                          setConfigUpdating(false);
                        }}
                      />
                    )}
                  </div>
                ),
              }}
              empty={{
                is: Object.keys(modifiedItems).length === 0,
                title: 'This is where you’ll manage your projects.',
                content: (
                  <p>
                    You can create a new project and manage the listed project.
                  </p>
                ),
                action: {
                  content: 'Add new entry',
                  prefix: <Plus />,
                  // @ts-ignore
                  onClick: () => setShowHandleConfig({ data: modifiedItems }),
                },
              }}
            >
              <Tools />
              <Resources
                // @ts-ignore
                originalItems={originalItems}
                modifiedItems={modifiedItems}
                setModifiedItems={setModifiedItems}
                editItem={editItem}
                restoreItem={restoreItem}
                deleteItem={deleteItem}
              />
            </Wrapper>
            <Handle
              show={showHandleConfig}
              setShow={setShowHandleConfig}
              onSubmit={(/** @type {{ key: any; value: any; }} */ val) => {
                addItem({ key: val.key, val: val.value });
                // @ts-ignore
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
