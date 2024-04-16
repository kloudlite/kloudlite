import { useOutletContext } from '@remix-run/react';
import Popup from '~/components/molecule/popup';
import { CopyButton, DetailItem } from '~/console/components/commons';
import { IDomains } from '~/console/server/gql/queries/domain-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { IClusterContext } from '../../_layout';


const DomainDetailPopup = ({
  visible,
  setVisible,
  data,
}: {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  data: ExtractNodeType<IDomains>;
}) => {
  const { cluster } = useOutletContext<IClusterContext>();
  return (
    <Popup.Root show={visible} onOpenChange={setVisible}>
      <Popup.Header>Domain detail</Popup.Header>
      <Popup.Content>
        {data && (
          <div className="flex flex-col gap-3xl">
            <div className="flex flex-row items-center gap-xl">
              <DetailItem title="Name" value={data.displayName} />
              <DetailItem title="Domain name" value={data.domainName} />
            </div>
            <DetailItem
              title="CNAME"
              value={
                <CopyButton
                  title={cluster.spec?.publicDNSHost}
                  value={cluster.spec?.publicDNSHost || ''}
                />
              }
            />
            <div className="p-xl rounded bg-surface-basic-active">
              <DetailItem
                title="Description"
                value={
                  <div>
                    Please update your DNS record for <b>{data.domainName}</b>{' '}
                    by setting CNAME to <b>{cluster.spec?.publicDNSHost}</b>
                  </div>
                }
              />
            </div>
          </div>
        )}
      </Popup.Content>
    </Popup.Root>
  );
};

export default DomainDetailPopup;
