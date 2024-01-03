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
import { IConfig } from '~/console/server/gql/queries/config-queries';
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
import { ISecret } from '~/console/server/gql/queries/secret-queries';
import Handle, { updateSecret } from './handle';
import Resources from './resources';
import Tools from './tools';

export const handle = () => {
  return {
    navbar: constants.nan,
  };
};

export const loader = async (ctx: IRemixCtx) => {
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

const ConfigBody = ({ secret }: { secret: ISecret }) => {
  const [showHandleConfig, setShowHandleConfig] =
    useState<IShowDialog<IModifiedItem>>(null);

  const [originalItems, setOriginalItems] = useState<IConfigOrSecretData>({});
  const [modifiedItems, setModifiedItems] = useState<IModifiedItem>({});

  const [configUpdating, setConfigUpdating] = useState(false);
  const { account, cluster, project, scope, workspace } = useParams();
  const api = useConsoleApi();
  const context = useOutletContext();
  const reload = useReload();

  const [searchText, setSearchText] = useState('');

  useEffect(() => {
    setOriginalItems(secret.stringData);
  }, []);

  useEffect(() => {
    try {
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
    } catch {
      //
    }
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
    <>
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
                  setShowHandleConfig({
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
                    await updateSecret({
                      api,
                      context,
                      secret,
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
          title: 'This is where you’ll manage your secrets.',
          content: (
            <p>You can create a new project and manage the listed project.</p>
          ),
          action: {
            content: 'Add new entry',
            prefix: <Plus />,
            onClick: () =>
              setShowHandleConfig({ type: 'add', data: modifiedItems }),
          },
        }}
        tools={<Tools searchText={searchText} setSearchText={setSearchText} />}
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
      <Handle
        show={showHandleConfig}
        setShow={setShowHandleConfig}
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
          setShowHandleConfig(null);
        }}
      />
    </>
  );
};

const Config = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ secret }) => {
        return <ConfigBody secret={secret} />;
      }}
    </LoadingComp>
  );
};

export default Config;
