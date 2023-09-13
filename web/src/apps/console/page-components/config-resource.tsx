import { useParams } from '@remix-run/react';
import { useState } from 'react';
import { dayjs } from '~/components/molecule/dayjs';
import List from '~/console/components/list';
import ResourceExtraAction from '../components/resource-extra-action';
import { IConfig } from '../server/gql/queries/config-queries';
import { parseFromAnn, parseName } from '../server/r-utils/common';
import { keyconstants } from '../server/r-utils/key-constants';

interface IConfigResource {
  onDelete: (item: any) => void;
  hasActions?: boolean;
  onClick?: (item: any) => void;
  linkComponent?: any;
  items: IConfig[];
}

const ConfigResource = ({
  items = [],
  onDelete,
  hasActions = true,
  onClick = (_) => _,
  linkComponent = null,
}: IConfigResource) => {
  const { account, cluster, project, scope, workspace } = useParams();
  const [selected, setSelected] = useState('');
  let props = {};
  if (linkComponent) {
    props = { linkComponent };
  }

  return (
    <List.Root {...props}>
      {items.map((item) => {
        const { name, entries, lastupdated } = {
          name: parseName(item),
          entries: [`${Object.keys(item?.data).length || 0} Entries`],
          lastupdated: (
            <span
              title={
                parseFromAnn(item, keyconstants.author)
                  ? `Updated By ${parseFromAnn(
                      item,
                      keyconstants.author
                    )}\nOn ${dayjs(item.updateTime).format('LLL')}`
                  : undefined
              }
            >
              {dayjs(item.updateTime).fromNow()}
            </span>
          ),
        };

        return (
          <List.Row
            onClick={() => {
              onClick(item);
              setSelected(name);
            }}
            pressed={selected === name}
            key={name}
            className="!p-3xl"
            to={
              linkComponent !== null
                ? `/${account}/${cluster}/${project}/${scope}/${workspace}/config/${name}`
                : undefined
            }
            columns={[
              {
                key: 1,
                className: 'flex-1',
                render: () => (
                  <div className="flex flex-col gap-sm">
                    <div className="bodyMd-semibold text-text-default">
                      {name}
                    </div>
                    <div className="bodySm text-text-soft">{lastupdated}</div>
                  </div>
                ),
              },
              {
                key: 2,
                render: () => (
                  <div className="text-text-soft bodyMd w-[140px] text-right">
                    {entries}
                  </div>
                ),
              },
              ...[
                ...(hasActions
                  ? [
                      {
                        key: 3,
                        render: () => <ResourceExtraAction options={[]} />,
                      },
                    ]
                  : []),
              ],
            ]}
          />
        );
      })}
    </List.Root>
  );
};

export default ConfigResource;
