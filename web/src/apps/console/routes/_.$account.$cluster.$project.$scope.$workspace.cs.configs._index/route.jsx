import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import {
  Link,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import logger from '~/root/lib/client/helpers/log';
import {
  getPagination,
  getSearch,
  parseDisplayname,
  parseName,
  parseNodes,
} from '~/console/server/r-urils/common';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { useLog } from '~/root/lib/client/hooks/use-log';
import ResourceList from '../../components/resource-list';
import Resources from './resources';
import Tools from './tools';
import HandleConfig from './handle-config';

const Configs = () => {
  const [showHandleConfig, setHandleConfig] = useState(null);
  const [showDeleteConfig, setShowDeleteConfig] = useState(false);

  const data = useOutletContext();
  useLog(data);

  const { promise } = useLoaderData();

  const { account, cluster, project, scope, workspace } = useParams();

  return (
    <>
      <LoadingComp data={promise}>
        {({ configsData }) => {
          const configs = parseNodes(configsData);
          return (
            <Wrapper
              empty={{
                is: configs.length === 0,
                title: 'This is where you’ll manage your Config.',
                content: (
                  <p>
                    You can create a new config and manage the listed configs.
                  </p>
                ),
                action: {
                  content: 'Create config',
                  prefix: Plus,
                  LinkComponent: Link,
                  onClick: () => {
                    setHandleConfig({ type: 'add', data: null });
                  },
                },
              }}
            >
              <div className="flex flex-col">
                <Tools />
              </div>
              {/* <List /> */}
              <ResourceList mode="list" linkComponent={Link} prefetchLink>
                {configs.map((d) => (
                  <ResourceList.ResourceItem
                    key={parseName(d)}
                    textValue={parseDisplayname(d)}
                    to={`/${account}/${cluster}/${project}/${scope}/${workspace}/config/${parseName(
                      d
                    )}`}
                  >
                    <Resources
                      item={d}
                      onDelete={(item) => {
                        setShowDeleteConfig(item);
                      }}
                    />
                  </ResourceList.ResourceItem>
                ))}
              </ResourceList>
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleConfig show={showHandleConfig} setShow={setHandleConfig} />
      {/* Alert Dialog for deleting config */}
      <AlertDialog
        show={showDeleteConfig}
        setShow={setShowDeleteConfig}
        title="Delete config"
        message={"Are you sure you want to delete 'kloud-root-ca.crt"}
        type="critical"
        okText="Delete"
        onSubmit={() => {}}
      />
    </>
  );
};

export default Configs;

export const handle = {
  subheaderAction: () => <Button content="Add new config" prefix={Plus} />,
};

export const loader = async (ctx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { project, scope, workspace } = ctx.params;

  const promise = pWrapper(async () => {
    try {
      const { data, errors } = await GQLServerHandler(ctx.request).listConfigs({
        project: {
          value: project,
          type: 'name',
        },
        scope: {
          value: workspace,
          type: scope === 'workspace' ? 'workspaceName' : 'environmentName',
        },
        pagination: getPagination(ctx),
        search: getSearch(ctx),
      });
      if (errors) {
        throw errors[0];
      }
      console.log(data);
      return { configsData: data };
    } catch (err) {
      logger.error(err);
      return { error: err.message };
    }
  });

  return defer({ promise });
};
