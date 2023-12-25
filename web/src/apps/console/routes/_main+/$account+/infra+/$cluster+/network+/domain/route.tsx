import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import logger from '~/root/lib/client/helpers/log';
import { IRemixCtx } from '~/root/lib/types/common';
import fake from '~/root/fake-data-generator/fake';
import DomainResources from './domain-resources';
import HandleDomain from './handle-domain';
import Tools from './tools';

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { data, errors } = await GQLServerHandler(ctx.request).listDomains({
      pagination: getPagination(ctx),
      search: getSearch(ctx),
    });
    if (errors) {
      logger.error(errors[0]);
      throw errors[0];
    }

    return {
      domainData: data || {},
    };
  });

  return defer({ promise });
};

const Domain = () => {
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();
  return (
    <>
      <LoadingComp
        data={promise}
        skeletonData={{
          domainData: fake.ConsoleListDomainsQuery.infra_listDomainEntries,
        }}
      >
        {({ domainData }) => {
          const domains = domainData.edges?.map(({ node }) => node);

          return (
            <Wrapper
              secondaryHeader={{
                title: 'Domain',
                action: domains.length > 0 && (
                  <Button
                    content="Add domain"
                    prefix={<Plus />}
                    variant="primary"
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                is: domains.length === 0,
                action: {
                  content: 'Add domain',
                  prefix: <Plus />,
                  variant: 'primary',
                  onClick: () => {
                    setVisible(true);
                  },
                },
                title: 'This is where you’ll oversees and control your domain.',
                content: (
                  <p>
                    You can add a new domain and exercise control over the
                    domains listed.
                  </p>
                ),
              }}
              tools={<Tools />}
            >
              <DomainResources items={domains} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleDomain
        {...{
          isUpdate: false,
          visible,
          setVisible,
        }}
      />
    </>
  );
};

export default Domain;
