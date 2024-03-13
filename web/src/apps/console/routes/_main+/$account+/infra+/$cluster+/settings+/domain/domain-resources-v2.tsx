import { useState } from 'react';
import { generateKey } from '~/components/utils';
import {
  ListBody,
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import { IDomains } from '~/console/server/gql/queries/domain-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import { useOutletContext } from '@remix-run/react';
import { IClusterContext } from '~/console/routes/_main+/$account+/infra+/$cluster+/_layout';
import ListV2 from '~/console/components/listV2';
import { Domain } from '~/console/components/icons';
import DomainDetailPopup from './domain-detail';

const RESOURCE_NAME = 'domain';
type BaseType = ExtractNodeType<IDomains>;

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: item.id,
    domainName: item.domainName,
    updateInfo: {
      author: `Updated by ${parseUpdateOrCreatedBy(item)}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

type OnAction = ({
  action,
  item,
}: {
  action: 'detail';
  item: BaseType;
}) => void;

interface IResource {
  items: BaseType[];
  onAction: OnAction;
}

const GridView = ({ items, onAction }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, domainName, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            onClick={() => onAction({ action: 'detail', item })}
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => <ListTitle title={name} />,
              },
              {
                key: generateKey(keyPrefix, domainName),
                render: () => <ListBody data={domainName} />,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                render: () => (
                  <ListItem
                    data={updateInfo.author}
                    subtitle={updateInfo.time}
                  />
                ),
              },
            ]}
          />
        );
      })}
    </Grid.Root>
  );
};

const ListView = ({ items, onAction }: IResource) => {
  return (
    <ListV2.Root
      data={{
        headers: [
          {
            render: () => 'Name',
            name: 'name',
            className: 'min-w-[180px] flex-1',
          },
          {
            render: () => 'Updated',
            name: 'updated',
            className: 'w-[180px]',
          },
        ],
        rows: items.map((i) => {
          const { domainName, updateInfo } = parseItem(i);

          return {
            columns: {
              name: {
                render: () => <ListTitle title={domainName} />,
              },
              updated: {
                render: () => (
                  <ListItem
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
            },
            onClick: () => onAction({ action: 'detail', item: i }),
          };
        }),
      }}
    />
  );
};

const DomainResourcesV2 = ({ items = [] }: { items: BaseType[] }) => {
  const [domainDetail, setDomainDetail] = useState<BaseType | null>(null);

  const { account } = useOutletContext<IAccountContext>();
  const { cluster } = useOutletContext<IClusterContext>();
  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.cluster:${parseName(
        cluster
      )}.domain_entries:${i.domainName}`;
    })
  );

  const props: IResource = {
    items,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'detail':
          setDomainDetail(item);
          break;
        default:
      }
    },
  };
  return (
    <>
      <ListGridView
        listView={<ListView {...props} />}
        gridView={<GridView {...props} />}
      />
      <DomainDetailPopup
        {...{
          visible: !!domainDetail,
          setVisible: () => setDomainDetail(null),
          data: domainDetail!,
        }}
      />
    </>
  );
};

export default DomainResourcesV2;
