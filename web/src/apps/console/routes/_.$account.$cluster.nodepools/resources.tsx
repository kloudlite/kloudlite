import { PencilLine } from '@jengaicons/react';

import Tooltip from '~/components/atoms/tooltip';
import { cn } from '~/components/utils';
import {
  ListBody,
  ListTitle,
} from '~/console/components/console-list-components';
import List from '~/console/components/list';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import {
  // INodepool,
  INodepools,
} from '~/console/server/gql/queries/nodepool-queries';
import {
  ExtractNodeType,
  // parseFromAnn,
  parseName,
} from '~/console/server/r-utils/common';
// import { keyconstants } from '~/console/server/r-utils/key-constants';
// import { findNodePlan, provisionTypes } from './nodepool-utils';

const NodeStatus = ({ nodes = [] }: { nodes: any }) => (
  <div className="flex flex-row gap-xl">
    <div className="flex flex-row gap-lg">
      {nodes?.map((node: any) => (
        <Tooltip.Root
          content={
            <div className="flex flex-col bodySm-medium text-text-strong">
              <span>
                <span className="text-text-default">{node.name}: </span>
                Error
              </span>
              <span>{node.ip}</span>
            </div>
          }
          key={node.id}
        >
          <div
            className={cn('w-2xl h-2xl', {
              'bg-icon-success': node.status === 'running',
              'bg-icon-warning': node.status === 'starting',
              'bg-icon-critical': node.status === 'stopped',
            })}
          />
        </Tooltip.Root>
      ))}
    </div>
    <span className="bodySm text-text-soft">
      {nodes.filter((node: any) => node.status === 'running').length}/
      {nodes.length} ready
    </span>
  </div>
);

const Resources = ({
  items = [],
  onEdit,
}: {
  items: ExtractNodeType<INodepools>[];
  onEdit: (item: ExtractNodeType<INodepools>) => void;
}) => {
  return (
    <List.Root>
      {items.map((item) => {
        return (
          <List.Row
            key={parseName(item)}
            className="!p-3xl"
            columns={[
              {
                key: 1,
                className: 'flex-1',
                render: () => (
                  <div className="flex flex-col gap-3xl">
                    <div className="flex flex-row items-center gap-3xl">
                      <ListTitle title="Agenpllo" className="flex-1" />
                      {/* <ListBody */}
                      {/*   data={ */}
                      {/*     provisionTypes.find( */}
                      {/*       (pt) => */}
                      {/*         pt.value === */}
                      {/*         item.spec.awsNodeConfig?.provisionMode */}
                      {/*     )?.label */}
                      {/*   } */}
                      {/*   className="w-[120px] text-right" */}
                      {/* /> */}
                      {/* <ListBody */}
                      {/*   data={`${item.spec.minCount} min - ${item.spec.maxCount} max`} */}
                      {/*   className="w-[120px] text-right underline" */}
                      {/* /> */}
                      {/* <ListBody */}
                      {/*   data={ */}
                      {/*     findNodePlan( */}
                      {/*       parseFromAnn(item, keyconstants.node_type) */}
                      {/*     )?.label */}
                      {/*   } */}
                      {/*   className="w-[160px] text-right" */}
                      {/* /> */}
                      <ListBody
                        data="4/8 running"
                        className="w-[100px] text-right"
                      />
                      <ListBody
                        data="5/18/23, 3:30 PM"
                        className="w-[140px] text-right"
                      />
                      <ResourceExtraAction
                        options={[
                          {
                            key: '1',
                            label: 'Edit',
                            icon: <PencilLine size={16} />,
                            type: 'item',
                            onClick: () => onEdit(item),
                          },
                        ]}
                      />
                    </div>
                    <NodeStatus nodes={[]} />
                  </div>
                ),
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

export default Resources;
