import {
  Trash,
  CheckCircleFill,
  Checks,
  CircleNotch,
  Circle,
  CircleFill,
  WarningCircleFill,
} from '~/console/components/icons';
import Tooltip from '~/components/atoms/tooltip';
import {
  Github__Com___Kloudlite___Api___Pkg___Types__SyncState as ISyncState,
  Github__Com___Kloudlite___Api___Pkg___Types__SyncAction as ISyncAction,
  Github__Com___Kloudlite___Operator___Pkg___Operator__CheckMetaIn as ICheckList,
} from '~/root/src/generated/gql/server';
import { Badge } from '~/components/atoms/badge';

interface IStatusMetaV2 {
  recordVersion: number;
  markedForDeletion?: boolean;
  status?: {
    checks?: {
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
    };
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

type IStatusViewType = 'full' | 'minimal';

type OverallStates = 'idle' | 'in-progress' | 'error' | 'ready' | 'deleting';

const state = ({
  state,
  type,
}: {
  state: OverallStates;
  type: IStatusViewType;
}) => {
  const iconSize = 16;
  const textData = () => {
    switch (state) {
      case 'in-progress':
        return 'In progress';
      case 'error':
        return 'Error';
      case 'ready':
        return 'Ready';
      case 'deleting':
        return 'Deleting';
      case 'idle':
      default:
        return 'Waiting';
    }
  };
  switch (state) {
    case 'ready':
      return {
        component:
          type === 'minimal' ? (
            <span className="text-text-primary">
              <Checks size={16} />
            </span>
          ) : (
            <Badge icon={<Checks />} type="info">
              {textData()}
            </Badge>
          ),
        text: textData(),
      };
    case 'in-progress':
      const progressComponent = (
        <span className="animate-spin relative flex items-center justify-center text-text-warning">
          <CircleNotch size={type === 'minimal' ? iconSize : 12} />
          <span className="absolute">
            <CircleFill size={8} />
          </span>
        </span>
      );
      return {
        component:
          type === 'minimal' ? (
            progressComponent
          ) : (
            <Badge icon={progressComponent} type="warning">
              {textData()}
            </Badge>
          ),
        text: textData(),
      };
    case 'deleting':
      return {
        component:
          type === 'minimal' ? (
            <span className="text-text-critical">
              <Trash size={iconSize} />
            </span>
          ) : (
            <Badge icon={<Trash />} type="critical">
              {textData()}
            </Badge>
          ),
        text: textData(),
      };
    case 'error':
      return {
        component:
          type === 'minimal' ? (
            <span className="text-text-critical">!!</span>
          ) : (
            <Badge icon={<span className="px-xs">!!</span>} type="critical">
              {textData()}
            </Badge>
          ),
        text: textData(),
      };
    default:
      return {
        component:
          type === 'minimal' ? (
            <span className="text-text-warning">
              <CircleFill size={iconSize} />
            </span>
          ) : (
            <Badge icon={<CircleFill />} type="warning">
              {textData()}
            </Badge>
          ),
        text: textData(),
      };
  }
};

const parseOverallState = (item: IStatusMetaV2): OverallStates => {
  const { status, markedForDeletion, syncStatus } = item;

  const checks = status?.checks;
  const checkList = status?.checkList;

  if (markedForDeletion && syncStatus.action === 'DELETE') {
    return 'deleting';
  }

  if (!checks) {
    return 'idle';
  }

  const mainStatus = checkList?.reduce(
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

export const SyncStatusV2 = ({
  item,
  type,
}: {
  item: IStatusMetaV2;
  type?: IStatusViewType;
}) => {
  const parseStage = (check: OverallStates) => {
    const iconSize = 12;

    switch (check) {
      case 'in-progress':
        return {
          icon: (
            <span className="flex items-center justify-center text-text-warning">
              <span className="animate-spin">
                <CircleNotch size={iconSize} />
              </span>
            </span>
          ),
        };
      case 'error':
        return {
          icon: (
            <span className="text-text-critical">
              <WarningCircleFill size={iconSize} />
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

  const getProgressItems = (item: IStatusMetaV2) => {
    const { status } = item;

    const checks = status?.checks;
    const checkList = status?.checkList;

    if (!checks) {
      return [];
    }

    const items = checkList?.reduce(
      (acc, curr) => {
        const k = checks[curr.name];
        if (acc.progress === 'done') {
          acc.items.push({
            ...curr,
            result: 'idle',
          });
          return acc;
        }

        const res = ((): { value: OverallStates; progress: string } => {
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
        value: 'idle' as OverallStates,
        items: [] as ({ result: OverallStates } & ICheckList)[],
        progress: 'init',
      }
    );

    return items?.items;
  };

  const _data = {
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

  const k = parseOverallState(item);
  const ic = getProgressItems(item);

  return (
    <div>
      <Tooltip.Root
        align="center"
        className="!max-w-[300px]"
        content={
          <div className="p-md flex flex-col gap-lg">
            <div className="bodyMd-medium">
              {state({ state: k, type: type || 'full' }).text}
            </div>
            <div className="flex flex-col gap-lg">
              {k === 'idle' && (
                <div className="bodySm">
                  Please wait while we are operating on this resource
                </div>
              )}

              {ic?.map((cl) => (
                <div
                  key={cl.name}
                  className="bodySm flex flex-row gap-xl items-center"
                >
                  <span>{parseStage(cl.result).icon}</span>
                  <span>{cl.title}</span>
                </div>
              ))}
            </div>
          </div>
        }
      >
        <div className="cursor-pointer">
          {
            state({
              state: k,
              type: type || 'full',
            }).component
          }
        </div>
      </Tooltip.Root>
    </div>
  );
};

export const status = ({ item }: { item: IStatusMetaV2 }) => {
  return parseOverallState(item);
};
