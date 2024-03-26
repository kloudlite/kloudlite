import { useState } from 'react';
import { cn, generateKey } from '~/components/utils';
import {
  ListBody,
  ListItem,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import Grid from '~/console/components/grid';
import ListGridView from '~/console/components/list-grid-view';
import {
  ExtractNodeType,
  parseName,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import { useOutletContext, useParams } from '@remix-run/react';
import { IBuildRuns } from '~/console/server/gql/queries/build-run-queries';
import AnimateHide from '~/components/atoms/animate-hide';
import { Button } from '~/components/atoms/button';
import {
  CheckCircleFill,
  GitBranch,
  PlayCircleFill,
  Tag,
  XCircleFill,
} from '@jengaicons/react';
import dayjs from 'dayjs';
import LogComp from '~/lib/client/components/logger';
import LogAction from '~/console/page-components/log-action';
import { useDataState } from '~/console/page-components/common-state';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import ListV2 from '~/console/components/listV2';
import { SyncStatusV2 } from '~/console/components/sync-status';
import { useWatchReload } from '~/lib/client/helpers/socket/useWatch';
import { IRepoContext } from '~/console/routes/_main+/$account+/repo+/$repo+/_layout';

const RESOURCE_NAME = 'build run';
type BaseType = ExtractNodeType<IBuildRuns>;

const parseItem = (item: BaseType) => {
  return {
    name: parseName(item),
    id: parseName(item),
    updateInfo: {
      time: parseUpdateOrCreatedOn(item),
    },
  };
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
                  <ListTitle
                    title={name}
                    subtitle={id}
                    // action={
                    //   <ExtraButton
                    //     onDelete={() => {
                    //       onDelete(item);
                    //     }}
                    //   />
                    // }
                  />
                ),
              },
              {
                key: generateKey(keyPrefix, 'time'),
                render: () => (
                  <ListBody data={`Last Updated ${updateInfo.time}`} />
                ),
              },
            ]}
          />
        );
      })}
    </Grid.Root>
  );
};

const ListItem_ = ({ item }: { item: BaseType }) => {
  const [open, setOpen] = useState<boolean>(false);
  const { account } = useOutletContext<IAccountContext>();

  if (item.metadata && !item.metadata.annotations) {
    item.metadata.annotations = {};
  }

  const commitHash = item.metadata?.annotations?.['github.com/commit'];

  // eslint-disable-next-line no-nested-ternary
  const state: 'running' | 'done' | 'error' = item.status?.isReady
    ? 'done'
    : item.status?.message?.RawMessage
    ? 'error'
    : 'running';

  const isLatest = dayjs(item.updateTime).isAfter(dayjs().subtract(3, 'hour'));

  const { state: st } = useDataState<{
    linesVisible: boolean;
    timestampVisible: boolean;
  }>('logs');

  return (
    <div className="flex flex-col flex-1">
      <div className="flex flex-row justify-between items-center gap-6xl">
        <div className="flex justify-between items-center flex-1">
          <div className="flex gap-xl items-center justify-start flex-1">
            <div>
              <span
                className={cn({
                  'text-text-success': state === 'done',
                  'text-text-critical': state === 'error',
                  'text-text-warning': state === 'running',
                })}
                title={
                  // eslint-disable-next-line no-nested-ternary
                  state === 'done'
                    ? 'Build completed successfully'
                    : state === 'error'
                    ? 'Build failed'
                    : 'Build in progress'
                }
              >
                {state === 'done' && (
                  <CheckCircleFill size={16} color="currentColor" />
                )}

                {state === 'error' && (
                  <XCircleFill size={16} color="currentColor" />
                )}

                {state === 'running' && (
                  <PlayCircleFill size={16} color="currentColor" />
                )}
              </span>
            </div>
            <ListTitle
              title={
                <div className="flex items-center gap-xl">
                  {item.metadata?.annotations?.['github.com/repository'] || ''}
                </div>
              }
              subtitle={
                <div className="flex items-center gap-xl pt-md">
                  <div>
                    {`#${commitHash?.substring(
                      commitHash.length - 7,
                      commitHash.length
                    )}`}
                  </div>
                  <div className="flex items-center gap-md">
                    <GitBranch size={12} />
                    {item.metadata?.annotations?.['github.com/branch'] || ''}
                  </div>

                  <div className="flex items-center gap-md">
                    {item.spec?.registry.repo.tags.map((tag) => (
                      <div className="flex items-center gap-md" key={tag}>
                        <Tag size={12} />
                        {tag}{' '}
                      </div>
                    ))}
                  </div>
                </div>
              }
            />
          </div>
        </div>

        <div className="bodyMd text-text-soft truncate">
          {parseUpdateOrCreatedOn(item)}
        </div>

        {/* <pre>{JSON.stringify(item, null, 2)}</pre> */}
        {isLatest && (
          <Button
            size="sm"
            variant="basic"
            content={open ? 'Hide Logs' : 'Show Logs'}
            onClick={() => setOpen((s) => !s)}
          />
        )}
      </div>

      <AnimateHide show={open} className="w-full pt-4xl">
        <LogComp
          {...{
            dark: true,
            width: '100%',
            height: '40rem',
            title: 'Logs',
            hideLineNumber: !st.linesVisible,
            hideTimestamp: !st.timestampVisible,
            actionComponent: <LogAction />,
            websocket: {
              account: parseName(account),
              cluster: item.clusterName,
              trackingId: item.id,
            },
          }}
        />
      </AnimateHide>
    </div>
  );
};

const ListView = ({ items }: { items: BaseType[] }) => {
  const [open, setOpen] = useState<string>('');

  const { account } = useParams();

  const { state: st } = useDataState<{
    linesVisible: boolean;
    timestampVisible: boolean;
  }>('logs');

  return (
    <ListV2.Root
      data={{
        headers: [
          {
            render: () => (
              <div className="flex flex-row">
                <span className="w-[48px]" />
                Name
              </div>
            ),
            name: 'name',
            className: 'w-[180px]',
          },
          {
            render: () => '',
            name: 'logs',
            className: 'min-w-[180px] flex-1 flex items-center justify-center',
          },
          {
            render: () => 'Status',
            name: 'status',
            className: 'min-w-[30px] flex items-center justify-center flex-1',
          },
          {
            render: () => 'Updated',
            name: 'updated',
            className: 'w-[180px]',
          },
        ],
        rows: items.map((item) => {
          const { name, updateInfo } = parseItem(item);

          if (item.metadata && !item.metadata.annotations) {
            item.metadata.annotations = {};
          }

          const isLatest = dayjs(item.updateTime).isAfter(
            dayjs().subtract(3, 'hour')
          );

          const commitHash = item.metadata?.annotations?.['github.com/commit'];

          return {
            columns: {
              name: {
                render: () => (
                  <div className="flex flex-col">
                    <ListTitle title={name} />

                    <div className="flex items-center gap-xl pt-md bodySm text-text-soft pulsable">
                      <div>
                        {`#${commitHash?.substring(
                          commitHash.length - 7,
                          commitHash.length
                        )}`}
                      </div>
                      <div className="flex items-center gap-md">
                        <GitBranch size={12} />
                        {item.metadata?.annotations?.['github.com/branch'] ||
                          ''}
                      </div>

                      <div className="flex items-center gap-md">
                        {item.spec?.registry.repo.tags.map((tag) => (
                          <div className="flex items-center gap-md" key={tag}>
                            <Tag size={12} />
                            {tag}{' '}
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                ),
              },

              logs: {
                render: () =>
                  isLatest ? (
                    <Button
                      size="sm"
                      variant="basic"
                      content={open === item.id ? 'Hide Logs' : 'Show Logs'}
                      onClick={(e) => {
                        e.preventDefault();

                        setOpen((s) => {
                          if (s === item.id) {
                            return '';
                          }
                          return item.id;
                        });
                      }}
                    />
                  ) : null,
              },
              status: {
                render: () => <SyncStatusV2 item={item} />,
              },
              updated: {
                render: () => <ListItem data={updateInfo.time} />,
              },
            },

            detail: (
              <AnimateHide
                onClick={(e) => e.preventDefault()}
                show={open === item.id}
                className="w-full"
              >
                <div className="w-full flex pb-2xl justify-center items-center pt-4xl">
                  <LogComp
                    {...{
                      dark: true,
                      width: '100%',
                      height: '40rem',
                      title: 'Logs',
                      hideLineNumber: !st.linesVisible,
                      hideTimestamp: !st.timestampVisible,
                      actionComponent: <LogAction />,
                      websocket: {
                        account: account || '',
                        cluster: item.clusterName,
                        trackingId: item.id,
                      },
                    }}
                  />
                </div>
              </AnimateHide>
            ),
            hideDetailSeperator: true,
          };
        }),
      }}
    />
  );
};

const BuildRunResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const { account } = useOutletContext<IAccountContext>();
  const { repoName } = useOutletContext<IRepoContext>();

  useWatchReload(
    items.map((i) => {
      return `account:${parseName(account)}.repo:${repoName}.build-run:${i.id}`;
    })
  );

  const props: IResource = {
    items,
    // onDelete: (item) => {
    //   setShowDeleteDialog(item);
    // },
  };

  return (
    <>
      <ListGridView
        listView={<ListView {...props} />}
        gridView={<GridView {...props} />}
      />
      <DeleteDialog
        resourceName={parseName(showDeleteDialog)}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          // try {
          //   const { errors } = await api.deleteVpnDevice({
          //     deviceName: parseName(showDeleteDialog),
          //     clusterName: params.cluster || '',
          //   });
          //
          //   if (errors) {
          //     throw errors[0];
          //   }
          //   reloadPage();
          //   toast.success(`${titleCase(RESOURCE_NAME)} deleted successfully`);
          //   setShowDeleteDialog(null);
          // } catch (err) {
          //   handleError(err);
          // }
        }}
      />
    </>
  );
};

export default BuildRunResources;
