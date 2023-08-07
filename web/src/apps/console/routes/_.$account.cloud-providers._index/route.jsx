import { useEffect, useState } from 'react';
import { Link, useLoaderData } from '@remix-run/react';
import { Plus, PlusFill } from '@jengaicons/react';
import { Button } from '~/components/atoms/button.jsx';
import logger from '~/root/lib/client/helpers/log';
import { useLog } from '~/root/lib/client/hooks/use-log';
import Wrapper from '~/console/components/Wrapper';
import ResourceList from '../../components/resource-list';
import { GQLServerHandler } from '../../server/gql/saved-queries';
import { dummyData } from '../../dummy/data';
import {
  getPagination,
  parseName,
  parseUpdationTime,
} from '../../server/r-urils/common';
import Tools from './Tools';
import Filters from './Filters';
import Resources from './Resources';
import HandleProvider from './HandleProvider';

const CloudProvidersIndex = () => {
  const [appliedFilters, setAppliedFilters] = useState(
    dummyData.appliedFilters
  );
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(15);
  const [totalItems, setTotalItems] = useState(100);
  const [viewMode, setViewMode] = useState('list');
  const [showAddProvider, setShowAddProvider] = useState(false);
  const { providerSecrets } = useLoaderData();

  useLog(providerSecrets);

  const data = providerSecrets?.edges?.map(({ node }) => node) || [];

  return (
    <>
      <Wrapper
        header={{
          title: 'Cloud Provider',
          action: data.length > 0 && (
            <Button
              variant="primary"
              content="Create Cloud Provider"
              prefix={PlusFill}
              onClick={() => {
                setShowAddProvider({ type: 'add', data: null });
              }}
            />
          ),
        }}
        empty={{
          is: data.length === 0,
          title:
            'This is the place where you will oversees the Cloud Provider.',
          content: (
            <p>
              You have the option to include a new Cloud Provider and oversee
              the existing Cloud Provider.
            </p>
          ),
          action: {
            content: 'Create Cloud Provider',
            prefix: Plus,
            LinkComponent: Link,
            onClick: () => {
              setShowAddProvider({ type: 'add', data: null });
            },
          },
        }}
        pagination={{
          currentPage,
          itemsPerPage,
          totalItems,
        }}
      >
        <div className="flex flex-col">
          <Tools viewMode={viewMode} setViewMode={setViewMode} />
          <Filters
            appliedFilters={appliedFilters}
            setAppliedFilters={setAppliedFilters}
          />
        </div>
        <ResourceList mode={viewMode}>
          {data.map((secret) => (
            <ResourceList.ResourceItem
              key={parseUpdationTime(secret) + parseName(secret)}
              textValue={parseUpdationTime(secret) + parseName(secret)}
            >
              <Resources
                item={secret}
                onEdit={(e) => {
                  setShowAddProvider({ type: 'edit', data: e });
                }}
              />
            </ResourceList.ResourceItem>
          ))}
        </ResourceList>
      </Wrapper>

      {/* Popup dialog for adding cloud provider */}
      <HandleProvider show={showAddProvider} setShow={setShowAddProvider} />
    </>
  );
};

export const loader = async (ctx) => {
  const { data, errors } = await GQLServerHandler(
    ctx.request
  ).listProviderSecrets({
    pagination: getPagination(ctx),
  });

  if (errors) {
    logger.error(errors);
  }

  return {
    providerSecrets: data,
  };
};

export default CloudProvidersIndex;
