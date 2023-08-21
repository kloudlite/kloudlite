import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import Wrapper from '~/console/components/wrapper';
import ResourceList from '~/console/components/resource-list';
import { useParams, useLoaderData } from '@remix-run/react';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import {
  getScopeAndProjectQuery,
  parseName,
} from '~/console/server/r-urils/common';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { defer } from '@remix-run/node';
import Tools from './tools';
import Resources from './resources';

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
  // const [data, setData] = useState(dummyData.configs);
  const { promise } = useLoaderData();
  const { account, cluster, project, scope, workspace } = useParams();
  return (
    <LoadingComp data={promise}>
      {({ config }) => {
        const { data } = config;
        return (
          <Wrapper
            header={{
              title: 'kloud-root-ca.crt',
              backurl: `/${account}/${cluster}/${project}/${scope}/${workspace}/cs/configs`,
              action: data.length > 0 && (
                <Button
                  variant="basic"
                  content="Add new entry"
                  prefix={PlusFill}
                />
              ),
            }}
            empty={{
              is: data.length === 0,
              title: 'This is where you’ll manage your projects.',
              content: (
                <p>
                  You can create a new project and manage the listed project.
                </p>
              ),
              action: {
                content: 'Add new entry',
                prefix: Plus,
              },
            }}
          >
            <div className="flex flex-col">
              <Tools />
            </div>
            <ResourceList>
              {data?.data?.map((d) => (
                <ResourceList.ResourceItem
                  key={parseName(d)}
                  textValue={parseName(d)}
                >
                  <Resources item={d} />
                </ResourceList.ResourceItem>
              ))}
            </ResourceList>
          </Wrapper>
        );
      }}
    </LoadingComp>
  );
};

export default Config;
