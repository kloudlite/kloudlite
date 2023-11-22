import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IShowDialog } from '~/console/components/types.d';
import Wrapper from '~/console/components/wrapper';
import { IDomains } from '~/console/server/gql/queries/domain-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { getPagination, getSearch } from '~/console/server/utils/common';
import { DIALOG_TYPE } from '~/console/utils/commons';
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
  const [showHandleDomain, setShowHandleDomain] =
    useState<IShowDialog<ExtractNodeType<IDomains> | null>>(null);
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp
      data={promise}
      skeletonData={{
        domainData: fake.ConsoleListDomainsQuery.infra_listDomainEntries,
      }}
    >
      {({ domainData }) => {
        const domains = domainData.edges?.map(({ node }) => node);

        console.log(domains);

        return (
          <>
            <Wrapper
              secondaryHeader={{
                title: 'Domain',
                action: domains.length > 0 && (
                  <Button
                    content="Add domain"
                    prefix={<Plus />}
                    variant="primary"
                    onClick={() => {
                      setShowHandleDomain({
                        type: DIALOG_TYPE.ADD,
                        data: null,
                      });
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
                    setShowHandleDomain({ type: DIALOG_TYPE.ADD, data: null });
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
            <HandleDomain
              show={showHandleDomain}
              setShow={setShowHandleDomain}
            />
          </>
        );
      }}
    </LoadingComp>
  );
};

export default Domain;
