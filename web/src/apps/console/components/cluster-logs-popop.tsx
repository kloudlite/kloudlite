import Popup from '@kloudlite/design-system/molecule/popup';
import LogComp from '~/root/lib/client/components/logger';
import { Button } from '@kloudlite/design-system/atoms/button';
import { useParams } from '@remix-run/react';
import { ExtractNodeType, parseName } from '../server/r-utils/common';
import LogAction from '../page-components/log-action';
import { IClusters } from '../server/gql/queries/cluster-queries';
import { IByocClusters } from '../server/gql/queries/byok-cluster-queries';
import { useDataState } from '../page-components/common-state';

type BaseType = ExtractNodeType<IClusters> & { type: 'normal' };
type ByokBaseType = ExtractNodeType<IByocClusters> & { type: 'byok' };
type CombinedBaseType = BaseType | ByokBaseType;

export const ViewClusterLogs = ({
  show,
  setShow,
  item,
}: {
  show: boolean;
  setShow: () => void;
  item: CombinedBaseType;
}) => {
  const { account } = useParams();
  const { state } = useDataState<{
    linesVisible: boolean;
    timestampVisible: boolean;
  }>('logs');

  return (
    <Popup.Root onOpenChange={setShow} show={show} className="!w-[800px]">
      <Popup.Header>{`${item.metadata.name} Logs:`}</Popup.Header>
      <Popup.Content>
        <LogComp
          {...{
            hideLineNumber: !state.linesVisible,
            hideTimestamp: !state.timestampVisible,
            className: 'flex-1',
            dark: true,
            width: '100%',
            height: '40rem',
            title: 'Logs',
            websocket: {
              account: account || '',
              cluster: parseName(item),
              trackingId: item.id,
            },
            actionComponent: <LogAction />,
          }}
        />
      </Popup.Content>
      <Popup.Footer>
        <Button variant="primary-outline" content="close" onClick={setShow} />
      </Popup.Footer>
    </Popup.Root>
  );
};
