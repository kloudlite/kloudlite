import { Plus } from '~/iotconsole/components/icons';
import { defer } from '@remix-run/node';
import { Link, useLoaderData } from '@remix-run/react';
import { Button } from '~/components/atoms/button.jsx';
import {
  LoadingComp,
  pWrapper,
} from '~/iotconsole/components/loading-component';
import Wrapper from '~/iotconsole/components/wrapper';
import { getPagination, getSearch } from '~/iotconsole/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { ensureAccountSet } from '~/iotconsole/server/utils/auth-utils';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';
import { parseNodes } from '~/iotconsole/server/r-utils/common';
import { useState } from 'react';
import Tools from './tools';
import DeploymentResource from './deployment-resource';
import HandleDeployment from './handle-deployment';

export const loader = (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { project } = ctx.params;

    const { data, errors } = await GQLServerHandler(
      ctx.request
    ).listIotDeployments({
      pq: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      deploymentData: data || {},
    };
  });

  return defer({ promise });
};

const Deployments = () => {
  // return <Wip />;
  const [visible, setVisible] = useState(false);

  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp
        data={promise}
        // skeletonData={{
        //   projectsData: fake.ConsoleListProjectsQuery.core_listProjects as any,
        // }}
      >
        {({ deploymentData }) => {
          const deployments = parseNodes(deploymentData);

          return (
            <Wrapper
              header={{
                title: 'Deployments',
                action: deployments.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create Deployment"
                    prefix={<Plus />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                is: deployments.length === 0,
                title: 'This is where you’ll manage your deployment.',
                content: (
                  <p>
                    You can create a new deployment and manage the listed
                    deployments.
                  </p>
                ),
                action: {
                  content: 'Create new deployment',
                  prefix: <Plus />,
                  onClick: () => {
                    setVisible(true);
                  },
                  linkComponent: Link,
                },
              }}
              tools={<Tools />}
              // pagination={{
              //   pageInfo: deploymentData.pageInfo,
              // }}
            >
              <DeploymentResource items={deployments} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleDeployment
        {...{
          visible,
          setVisible,
          isUpdate: false,
        }}
      />
    </>
  );
};

export default Deployments;
