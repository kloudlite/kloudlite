import {
  ArrowsClockwise,
  Trash,
  Info,
  Warning,
  WarningCircle,
} from '@jengaicons/react';
import Tooltip from '~/components/atoms/tooltip';
import { titleCase } from '~/components/utils';
import {
  Github__Com___Kloudlite___Api___Pkg___Types__SyncState as SyncState,
  Github__Com___Kloudlite___Api___Pkg___Types__SyncAction as SyncAction,
} from '~/root/src/generated/gql/server';

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
    action: SyncAction;
    error?: string;
    lastSyncedAt?: any;
    recordVersion: number;
    state: SyncState;
    syncScheduledAt?: any;
  };
}

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

export type IStatus = 'deleting' | 'notready' | 'syncing' | 'none';
type IResourceType = 'nodepool';

export const parseStatus = ({
  item,
  type,
}: {
  item: IStatusMeta;
  type?: IResourceType;
}) => {
  let status: IStatus = 'none';

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
  key,
  className,
  item,
  type,
}: {
  key: string;
  className?: string;
  item: IStatusMeta;
  type?: IResourceType;
}) => {
  return {
    key,
    className,
    render: () => <SyncStatus item={item} />,
    status: parseStatus({ item, type }),
  };
};

export default SyncStatus;
