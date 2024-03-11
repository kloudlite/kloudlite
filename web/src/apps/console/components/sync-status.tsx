import {
  ArrowsClockwise,
  Trash,
  Info,
  WarningCircle,
  CheckCircleFill,
  Checks,
  CircleNotch,
  XCircleFill,
  Circle,
  CircleFill,
  X,
} from '~/console/components/icons';
import Tooltip from '~/components/atoms/tooltip';
import { titleCase } from '~/components/utils';
import {
  Github__Com___Kloudlite___Api___Pkg___Types__SyncState as ISyncState,
  Github__Com___Kloudlite___Api___Pkg___Types__SyncAction as ISyncAction,
  Github__Com___Kloudlite___Operator___Pkg___Operator__CheckMetaIn as ICheckList,
} from '~/root/src/generated/gql/server';
import { Badge } from '~/components/atoms/badge';

interface IStatusMeta {
  metadata?: { generation: number };
  recordVersion: number;
  markedForDeletion?: boolean;
  status?: {
    checks?: any;
    isReady: boolean;
    lastReadyGeneration?: number;
    lastReconcileTime?: any;
    message?: { RawMessage?: any };
    resources?: Array<{
      apiVersion: string;
      kind: string;
      name: string;
      namespace: string;
    }>;
  };
  syncStatus: {
    action: ISyncAction;
    error?: string;
    lastSyncedAt?: any;
    recordVersion: number;
    state: ISyncState;
    syncScheduledAt?: any;
  };
}

interface IStatusMetaV2 {
  metadata?: { generation: number };
  recordVersion: number;
  markedForDeletion?: boolean;
  status?: {
    checks?: any;
    isReady: boolean;
    lastReadyGeneration?: number;
    lastReconcileTime?: any;
    message?: { RawMessage?: any };
    resources?: Array<{
      apiVersion: string;
      kind: string;
      name: string;
      namespace: string;
    }>;
    checkList?: ICheckList[];
  };
  syncStatus: {
    action: ISyncAction;
    error?: string;
    lastSyncedAt?: any;
    recordVersion: number;
    state: ISyncState;
    syncScheduledAt?: any;
  };
}

const Stage = {
  Waiting: 'yet-to-be-reconciled',
  InProgress: 'under-reconcilation',
  Error: 'errored-during-reconcilation',
  Completed: 'finished-reconcilation',
};

type OverallStates = 'idle' | 'in-progress' | 'error' | 'ready' | 'deleting';

const State = ({ state }: { state: OverallStates }) => {
  switch (state) {
    case 'ready':
      return (
        <Badge icon={<Checks />} type="info">
          Ready
        </Badge>
      );
    case 'in-progress':
      return (
        <Badge
          icon={
            <span className="animate-spin relative flex items-center justify-center">
              <CircleNotch size={12} />
              <span className="absolute">
                <CircleFill size={8} />
              </span>
            </span>
          }
          type="warning"
        >
          In progress
        </Badge>
      );
    case 'deleting':
      return 'Deleting';
    case 'error':
      return (
        <Badge icon={<X />} type="critical">
          Error
        </Badge>
      );
    default:
      return (
        <Badge icon={<CircleFill />} type="warning">
          Waiting
        </Badge>
      );
  }
};

export const SyncStatusV2 = ({ item }: { item: IStatusMetaV2 }) => {
  const { status } = item;

  const parseStage = (check: OverallStates) => {
    const iconSize = 12;

    switch (check) {
      case 'in-progress':
        return {
          icon: (
            <span className="text-text-warning">
              <CircleNotch size={iconSize} />
            </span>
          ),
        };
      case 'error':
        return {
          icon: (
            <span className="text-text-critical">
              <XCircleFill size={iconSize} />
            </span>
          ),
        };
      case 'ready':
        return {
          icon: (
            <span className="text-text-success">
              <CheckCircleFill size={iconSize} />
            </span>
          ),
        };
      case 'idle':
      default:
        return {
          icon: (
            <span className="text-text-soft">
              <Circle size={iconSize} />
            </span>
          ),
        };
    }
  };

  const parseOverallState = (s: typeof status): OverallStates => {
    /*
    WaitingState   State = "yet-to-be-reconciled"
  RunningState   State = "under-reconcilation"
  ErroredState   State = "errored-during-reconcilation"
  CompletedState State = "finished-reconcilation"
    */

    const checks: {
      [key: string]: {
        error: string;
        generation: number;
        message: string;
        state:
          | 'yet-to-be-reconciled'
          | 'under-reconcilation'
          | 'errored-during-reconcilation'
          | 'finished-reconcilation';
        status: boolean;
      };
    } = s?.checks;
    if (!checks) {
      return 'idle';
    }

    /*
    if no one stared -> idle

    if any fail all fail

    if any not in progress -> in progress

    if all done -> done
    */

    const mainStatus = status?.checkList?.reduce(
      (acc, curr) => {
        const k = checks[curr.name];
        if (acc.progress === 'done') {
          return acc;
        }

        if (k) {
          if (acc.value === 'idle' && k.state === 'yet-to-be-reconciled') {
            return {
              value: 'idle',
              progress: 'done',
            };
          }

          if (k.state === 'under-reconcilation') {
            return {
              value: 'in-progress',
              progress: 'done',
            };
          }

          if (k.state === 'errored-during-reconcilation') {
            return {
              value: 'error',
              progress: 'done',
            };
          }

          if (k.state === 'finished-reconcilation') {
            return {
              value: 'ready',
              progress: 'init',
            };
          }
        }

        return acc;
      },
      {
        value: 'idle',
        progress: 'init',
      }
    );

    return (mainStatus?.value as OverallStates) || 'idle';
  };

  const getProgressItems = (s: typeof status) => {
    const checks: {
      [key: string]: {
        error: string;
        generation: number;
        message: string;
        state:
          | 'yet-to-be-reconciled'
          | 'under-reconcilation'
          | 'errored-during-reconcilation'
          | 'finished-reconcilation';
        status: boolean;
      };
    } = s?.checks;
    if (!checks) {
      return 'idle';
    }

    /*
    if no one stared -> idle

    if any fail all fail

    if any not in progress -> in progress

    if all done -> done
    */

    const items = status?.checkList?.reduce(
      (acc, curr) => {
        const k = checks[curr.name];
        if (acc.progress === 'done') {
          acc.items.push({
            ...curr,
            result: 'idle',
          });
          return acc;
        }

        const res = (() => {
          if (k) {
            if (acc.value === 'idle' && k.state === 'yet-to-be-reconciled') {
              return {
                value: 'idle',
                progress: 'done',
              };
            }

            if (k.state === 'under-reconcilation') {
              return {
                value: 'in-progress',
                progress: 'done',
              };
            }

            if (k.state === 'errored-during-reconcilation') {
              return {
                value: 'error',
                progress: 'done',
              };
            }

            if (k.state === 'finished-reconcilation') {
              return {
                value: 'ready',
                progress: 'init',
              };
            }
          }

          return acc;
        })();

        acc.items.push({
          ...curr,
          result: res?.value,
        });

        acc.value = res.value;
        acc.progress = res.progress;

        return acc;
      },
      {
        value: 'idle',
        items: [],
        progress: 'init',
      }
    );

    return items?.items;
  };

  const data = {
    checks: {
      'deployment-svc-and-hpa-created': {
        state: 'under-reconcilation',
        status: true,
      },
      'deployment-ready': {
        state: 'finished-reconcilation',
        status: false,
      },
    },
    checkList: [
      {
        title: 'Scaling configured',
        name: 'deployment-svc-and-hpa-created',
      },
      {
        title: 'Deployment ready',
        name: 'deployment-ready',
      },
    ],
  };

  const k = parseOverallState(data);

  const ic = getProgressItems(data);
  console.log(ic);
  return (
    <div>
      <Tooltip.Root
        align="center"
        className="!max-w-[300px]"
        content={
          <div className="p-md flex flex-col gap-lg">
            <div className="bodyMd-medium" />
            <div className="flex flex-col gap-lg">
              {ic.map((cl) => (
                <div
                  key={cl.name}
                  className="bodySm flex flex-row gap-xl items-center"
                >
                  <span>{parseStage(cl.name).icon}</span>
                  <span>{cl.title}</span>
                </div>
              ))}
            </div>
          </div>
        }
      >
        <div className="cursor-pointer">
          <State state={k} />
        </div>
      </Tooltip.Root>
    </div>
  );
};

const SyncStatus = ({ item }: { item: IStatusMeta }) => {
  const statusIconSize = 16;

  const getMessages = () => {
    let errors: Array<string> = [];
    if (item.status?.checks) {
      try {
        const err = Object.entries(item.status?.checks)
          .map(([_key, values]) => {
            const val = values as unknown as Record<string, any>;
            if (val.generation === item.metadata?.generation && !val.status) {
              return val.message;
            }
            return null;
          })
          .filter((er) => !!er);
        errors = [...err];
      } catch {
        //
      }
    }

    if (errors.length > 0) {
      return {
        errors,
        render: (
          <div className="pulsable text-text-strong">
            <Tooltip.Root
              content={
                <div className="flex flex-col gap-lg text-text-strong text-xs">
                  {errors.map((err) => (
                    <span key={err}>{err}</span>
                  ))}
                </div>
              }
            >
              <span className="animate-pulse">
                <Info size={statusIconSize} />
              </span>
            </Tooltip.Root>
          </div>
        ),
      };
    }

    return null;
  };

  if (item.markedForDeletion) {
    return (
      <div className="pulsable text-text-critical">
        <Tooltip.Root
          content={
            <div className="flex flex-col gap-lg overflow-hidden">
              <span className="bodySm-semibold">Deleting</span>
              {getMessages()?.errors.map((error) => (
                <span
                  key={error}
                  className="bodySm overflow-hidden break-words"
                >
                  {titleCase(error)}
                </span>
              ))}
            </div>
          }
        >
          <span className="animate-pulse">
            <Trash size={statusIconSize} />
          </span>
        </Tooltip.Root>
      </div>
    );
  }

  if (item.syncStatus.error) {
    return (
      <div className="pulsable text-text-critical">
        <Tooltip.Root
          content={
            <div className="flex flex-col gap-md">
              <span className="bodySm-semibold">Error</span>
              <span className="bodySm">{titleCase(item.syncStatus.error)}</span>
            </div>
          }
        >
          <span className="animate-pulse">
            <WarningCircle size={statusIconSize} />
          </span>
        </Tooltip.Root>
      </div>
    );
  }

  if (
    item.recordVersion !== item.syncStatus.recordVersion ||
    !item.status?.isReady
  ) {
    return (
      <div className="pulsable flex flex-row items-center gap-lg">
        <Tooltip.Root
          content={
            <div className="flex flex-col gap-lg">
              <span className="bodySm-semibold">Not ready</span>
              {getMessages()?.errors.map((error) => (
                <span key={error} className="bodySm">
                  {titleCase(error)}
                </span>
              ))}
            </div>
          }
        >
          <span className="animate-spin delay-300 text-text-disabled">
            <ArrowsClockwise size={statusIconSize} />
          </span>
        </Tooltip.Root>

        {/* {getError()} */}
      </div>
    );
  }

  return null;
};

export type IStatus = 'deleting' | 'notready' | 'syncing' | 'ready';
type IResourceType = 'nodepool';

export const parseStatus = ({
  item,
  type,
}: {
  item: IStatusMeta;
  type?: IResourceType;
}) => {
  let status: IStatus = 'ready';

  if (item.markedForDeletion) {
    status = 'deleting';
  } else if (!item.status?.isReady) {
    switch (type) {
      case 'nodepool':
        status = 'syncing';
        break;
      default:
        status = 'notready';
    }
  }

  return status;
};

export const listStatus = ({
  item,
  type,
}: {
  item: IStatusMeta;
  type?: IResourceType;
}) => {
  return {
    render: () => (
      <div className="min-w-[20px]">
        <SyncStatus item={item} />
      </div>
    ),
    status: parseStatus({ item, type }),
  };
};

export default SyncStatus;
