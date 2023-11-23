import { CodeSimple, PencilLine, Trash } from '@jengaicons/react';
import { generateKey, titleCase } from '~/components/utils';
import ConsoleAvatar from '~/console/components/console-avatar';
import {
  ListItemWithSubtitle,
  ListTitleWithSubtitleAvatar,
} from '~/console/components/console-list-components';
import Grid from '~/console/components/grid';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { INodepools } from '~/console/server/gql/queries/nodepool-queries';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useState } from 'react';
import { parseStatus } from '~/console/utils/commons';
import Popup from '~/components/molecule/popup';
import { HighlightJsLogs } from 'react-highlightjs-logs';
import { yamlDump } from '~/console/components/diff-viewer';
import DeleteDialog from '~/console/components/delete-dialog';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import HandleNodePool from './handle-nodepool';

const RESOURCE_NAME = 'nodepool';
type BaseType = ExtractNodeType<INodepools>;

const parseItem = (item: BaseType) => {
  return {
    name: item.displayName,
    id: parseName(item),
    updateInfo: {
      author: `Updated by ${titleCase(parseUpdateOrCreatedBy(item))}`,
      time: parseUpdateOrCreatedOn(item),
    },
  };
};

const ShowCodeInModal = ({
  text,
  visible,
  setVisible,
}: {
  text: string;
  visible: boolean;
  setVisible: (v: boolean) => void;
}) => {
  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      {/* <Popup.Header>Resource Yaml</Popup.Header> */}
      <Popup.Content className="!p-0">
        <HighlightJsLogs
          width="100%"
          height="40rem"
          title="Yaml Code"
          dark
          selectableLines
          text={text}
          language="yaml"
        />
      </Popup.Content>
    </Popup.Root>
  );
};

const ExtraButton = ({ item }: { item: BaseType }) => {
  const [visible, setVisible] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showHandleNodepool, setShowHandleNodepool] = useState(false);

  const reload = useReload();
  const api = useConsoleApi();

  return (
    <>
      <ResourceExtraAction
        options={[
          {
            key: '1',
            label: 'Edit',
            icon: <PencilLine size={16} />,
            type: 'item',
            onClick: () => setShowHandleNodepool(true),
          },
          {
            key: '2',
            label: 'Resource Yaml',
            icon: <CodeSimple size={16} />,
            type: 'item',
            onClick: () => setVisible(true),
          },
          {
            label: 'Delete',
            icon: <Trash size={16} />,
            type: 'item',
            onClick: () => setShowDeleteDialog(true),
            key: 'delete',
            className: '!text-text-critical',
          },
        ]}
      />
      <HandleNodePool
        {...{
          isUpdate: true,
          visible: !!showHandleNodepool,
          setVisible: () => {
            setShowHandleNodepool(false);
          },
          data: item,
        }}
      />

      <ShowCodeInModal
        visible={visible}
        text={yamlDump(item)}
        setVisible={setVisible}
      />

      <DeleteDialog
        resourceName={item.displayName}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteNodePool({
              clusterName: item.clusterName,
              poolName: parseName(item),
            });

            if (errors) {
              throw errors[0];
            }
            reload();
            toast.success(`${titleCase(RESOURCE_NAME)} is added for deletion.`);
            setShowDeleteDialog(false);
          } catch (err) {
            handleError(err);
          }
        }}
      />
    </>
  );
};
interface IResource {
  items: BaseType[];
}

const GridView = ({ items }: IResource) => {
  return (
    <Grid.Root className="!grid-cols-1 md:!grid-cols-3">
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <Grid.Column
            key={id}
            rows={[
              {
                key: generateKey(keyPrefix, name + id),
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    avatar={<ConsoleAvatar name={id} />}
                    title={name}
                    subtitle={id}
                    action={<ExtraButton item={item} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                render: () => (
                  <ListItemWithSubtitle
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

const ListView = ({ items }: IResource) => {
  return (
    <List.Root>
      {items.map((item, index) => {
        const { name, id, updateInfo } = parseItem(item);
        const keyPrefix = `${RESOURCE_NAME}-${id}-${index}`;
        return (
          <List.Row
            key={id}
            className="!p-3xl"
            columns={[
              {
                key: generateKey(keyPrefix, name + id),
                className: 'flex-1',
                render: () => (
                  <ListTitleWithSubtitleAvatar
                    title={name}
                    subtitle={id}
                    avatar={<ConsoleAvatar name={id} />}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'status'),
                className: 'w-[180px]',
                render: () => parseStatus(item).component,
              },
              {
                key: generateKey(keyPrefix, updateInfo.author),
                className: 'w-[180px]',
                render: () => (
                  <ListItemWithSubtitle
                    data={`${updateInfo.author}`}
                    subtitle={updateInfo.time}
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'action'),
                render: () => <ExtraButton item={item} />,
              },
            ]}
          />
        );
      })}
    </List.Root>
  );
};

const NodepoolResources = ({ items = [] }: { items: BaseType[] }) => {
  return (
    <ListGridView
      gridView={<GridView items={items} />}
      listView={<ListView items={items} />}
    />
  );
};

export default NodepoolResources;
