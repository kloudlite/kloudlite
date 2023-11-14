import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData, useOutletContext, useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import {
  IConfigOrSecretData,
  IModifiedItem,
  IShowDialog,
} from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  getScopeAndProjectQuery,
  parseName,
} from '~/console/server/r-utils/common';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { constants } from '~/console/server/utils/constants';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { IRemixCtx } from '~/root/lib/types/common';
import { ManageSecretDialog, updateSecret } from './handle';
import Resources from './resources';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
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

const DataSetter = ({ set = (_: any) => _, value }: any) => {
  useEffect(() => {
    set(value || {});
  }, [value]);
  return null;
};

const Secret = () => {
  const [showHandleSecret, setShowHandleSecret] =
    useState<IShowDialog<IModifiedItem>>(null);
  const [originalItems, setOriginalItems] = useState<IConfigOrSecretData>({});
  const [modifiedItems, setModifiedItems] = useState<IModifiedItem>({});
  const [secretUpdating, setSecretUpdating] = useState(false);
  const { promise } = useLoaderData<typeof loader>();
  const { account, cluster, project, scope, workspace } = useParams();

  const [searchText, setSearchText] = useState('');

  const api = useConsoleApi();
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

  const changesCount = () => {
    return Object.values(modifiedItems).filter(
      (mi) =>
        mi.delete ||
        mi.insert ||
        (mi.newvalue != null && mi.newvalue !== mi.value)
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
                        setShowHandleSecret({
                          type: 'Add',
                          data: modifiedItems,
                        })
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
                          if (!secret) {
                            return;
                          }
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
                  onClick: () =>
                    setShowHandleSecret({ type: '', data: modifiedItems }),
                },
              }}
              tools={
                <Tools searchText={searchText} setSearchText={setSearchText} />
              }
            >
              <Resources
                searchText={searchText.trim()}
                modifiedItems={modifiedItems}
                editItem={(item, value) => {
                  if (modifiedItems[item.key].insert) {
                    setModifiedItems((prev) => ({
                      ...prev,
                      [item.key]: { ...item.value, value },
                    }));
                  } else {
                    setModifiedItems((prev) => ({
                      ...prev,
                      [item.key]: { ...item.value, newvalue: value },
                    }));
                  }
                }}
                restoreItem={({ key }) => {
                  setModifiedItems((prev) => ({
                    ...prev,
                    [key]: {
                      value: originalItems[key],
                      delete: false,
                      insert: false,
                      newvalue: null,
                      edit: false,
                    },
                  }));
                }}
                deleteItem={(item) => {
                  if (originalItems[item.key]) {
                    setModifiedItems((prev) => ({
                      ...prev,
                      [item.key]: { ...item.value, delete: true, y: 'x' },
                    }));
                  } else {
                    const mItems = { ...modifiedItems };
                    delete mItems[item.key];
                    setModifiedItems(mItems);
                  }
                }}
              />
            </Wrapper>
            <ManageSecretDialog
              show={showHandleSecret}
              setShow={setShowHandleSecret}
              onSubmit={(val) => {
                setModifiedItems((prev) => ({
                  [val.key]: {
                    value: val.value,
                    insert: true,
                    delete: false,
                    edit: false,
                    newvalue: null,
                  },
                  ...prev,
                }));
                setShowHandleSecret(null);
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

export default Secret;
