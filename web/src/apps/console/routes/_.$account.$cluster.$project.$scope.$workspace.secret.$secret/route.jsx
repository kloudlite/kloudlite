import { Plus, PlusFill } from '@jengaicons/react';
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
import { Button } from '~/components/atoms/button';
import { getScopeAndProjectQuery } from '~/console/server/utils/common';
import Tools from './tools';
import Resources from './resources';
import Handle, { updateSecret } from './handle';

const DataSetter = ({ set = (_) => _, value }) => {
  useEffect(() => {
    set(value || {});
  }, [value]);
  return null;
};

const Secret = () => {
  const [showHandleSecret, setShowHandleSecret] = useState(null);
  const [originalItems, setOriginalItems] = useState({});
  const [modifiedItems, setModifiedItems] = useState({});
  const [secretUpdating, setSecretUpdating] = useState(false);
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

  const deleteItem = ({ key, value }) => {
    if (originalItems[key]) {
      setModifiedItems((prev) => ({
        ...prev,
        [key]: { ...value, delete: true },
      }));
    } else {
      const mItems = { ...modifiedItems };
      delete mItems[key];
      setModifiedItems(mItems);
    }
  };

  const editItem = ({ key, value }, val) => {
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

  const restoreItem = ({ key }) => {
    setModifiedItems((prev) => ({
      ...prev,
      [key]: {
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
        mi.delete || mi.insert || (mi.newvalue && mi.newvalue !== mi.value)
    ).length;
  };

  return (
    <LoadingComp data={promise}>
      {({ secret }) => {
        const { stringData: d } = secret;
        return (
          <>
            <DataSetter set={setOriginalItems} value={d} />
            <Wrapper
              header={{
                title: parseName(secret),
                backurl: `/${account}/${cluster}/${project}/${scope}/${workspace}/cs/secrets`,
                action: Object.keys(modifiedItems).length > 0 && (
                  <div className="flex flex-row items-center gap-lg">
                    <Button
                      variant="outline"
                      content="Add new entry"
                      prefix={<PlusFill />}
                      onClick={() =>
                        setShowHandleSecret({ data: modifiedItems })
                      }
                    />
                    {changesCount() > 0 && (
                      <Button variant="basic" content="Discard" />
                    )}
                    {changesCount() > 0 && (
                      <Button
                        variant="primary"
                        content={`Commit ${changesCount()} changes`}
                        loading={secretUpdating}
                        onClick={async () => {
                          setSecretUpdating(true);
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

                          await updateSecret({
                            api,
                            context,
                            secret,
                            data: k,
                            reload,
                          });
                          setSecretUpdating(false);
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
                  onClick: () => setShowHandleSecret({ data: modifiedItems }),
                },
              }}
            >
              <Tools />
              <Resources
                originalItems={originalItems}
                modifiedItems={modifiedItems}
                setModifiedItems={setModifiedItems}
                editItem={editItem}
                restoreItem={restoreItem}
                deleteItem={deleteItem}
              />
            </Wrapper>
            <Handle
              show={showHandleSecret}
              setShow={setShowHandleSecret}
              onSubmit={(val) => {
                addItem({ key: val.key, val: val.value });
                setShowHandleSecret(false);
              }}
            />
          </>
        );
      }}
    </LoadingComp>
  );
};

export const handle = () => {
  return {
    navbar: constants.nan,
  };
};

export const loader = async (ctx) => {
  // main promise
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    ensureClusterSet(ctx);

    const { secret } = ctx.params;

    const { data, errors } = await GQLServerHandler(ctx.request).getSecret({
      name: secret,
      ...getScopeAndProjectQuery(ctx),
    });

    if (errors) {
      throw errors[0];
    }
    return { secret: data };
  });

  return defer({ promise });
};

export default Secret;
