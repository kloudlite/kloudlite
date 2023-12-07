/* eslint-disable jsx-a11y/control-has-associated-label */
import { useOutletContext } from '@remix-run/react';
import { Box, BoxPrimitive } from '~/console/components/common-console-components';
import Wrapper from '~/console/components/wrapper';
import {
  parseName,
  parseUpdateOrCreatedBy,
  parseUpdateOrCreatedOn,
} from '~/console/server/r-utils/common';
import HighlightJsLog from '~/console/components/logger';
import { DownloadSimple } from '@jengaicons/react';
import { downloadFile, renderCloudProvider } from '~/console/utils/commons';
import { Chip } from '~/components/atoms/chips';
import { DetailItem } from '~/console/components/commons';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { handleError } from '~/root/lib/utils/common';
import { toast } from '~/components/molecule/toast';
import { Suspense, useState } from 'react';
import { TitleBox } from '~/console/components/raw-wrapper';
import { IClusterContext } from '../_.$account.$cluster';

// const KubeConfigDownload = ({ cluster }: { cluster: string }) => {
//   const filename = `${cluster} kubeconfig.yaml`;
//   const api = useConsoleApi();

//   const [loading, setLoading] = useState(false);

//   const downloadConfig = async () => {
//     setLoading(true);
//     try {
//       const { errors, data } = await api.getCluster({
//         name: cluster,
//       });
//       if (errors) {
//         throw errors[0];
//       }
//       if (data.adminKubeconfig) {
//         const { encoding, value } = data.adminKubeconfig;
//         let linkSource = '';
//         switch (encoding) {
//           case 'base64':
//             linkSource = atob(value);
//             break;
//           default:
//             linkSource = value;
//         }

//         downloadFile({
//           filename,
//           data: linkSource,
//           format: 'text/plain',
//         });
//       } else {
//         toast.error('Kubeconfig not found.');
//       }
//     } catch (err) {
//       handleError(err);
//     } finally {
//       setLoading(false);
//     }
//   };
//   return (
//     <Chip
//       type="CLICKABLE"
//       item={cluster}
//       label="Download"
//       prefix={<DownloadSimple />}
//       loading={loading}
//       onClick={() => {
//         downloadConfig();
//       }}
//     />
//   );
// };

const ClusterInfo = () => {
  const { cluster } = useOutletContext<IClusterContext>();

  const providerInfo = () => {
    const provider = cluster.spec?.cloudProvider;
    switch (provider) {
      case 'aws':
        return (
          <DetailItem title="Region" value={cluster.spec?.aws?.region || ''} />
        );
      default:
        return null;
    }
  };
  return (
    <Wrapper
      header={{
        title: 'Overview',
      }}
    >
      <div className="flex flex-col gap-6xl">
        <BoxPrimitive>
          <div className="flex flex-col">
            <div className="flex flex-row gap-3xl flex-wrap">
              <DetailItem title="Cluster ID" value={parseName(cluster)} />
              {/* {!!cluster.adminKubeconfig && (
                <DetailItem
                  title="Kube config"
                  value={<KubeConfigDownload cluster={parseName(cluster)} />}
                />
              )} */}

              <DetailItem
                title="Last updated"
                value={`By ${parseUpdateOrCreatedBy(
                  cluster
                )} ${parseUpdateOrCreatedOn(cluster)}`}
              />
              <DetailItem
                title="Availability mode"
                value={cluster.spec?.availabilityMode || ''}
              />
              <DetailItem
                title="Cluster Internal Dns Host"
                value={cluster.spec?.clusterInternalDnsHost || ''}
              />
              <DetailItem
                title="Cloudflare Enabled"
                value={
                  cluster.spec?.cloudflareEnabled ? 'Enabled' : 'Disabled' || ''
                }
              />
              <DetailItem
                title="Backup To S3 Enabled"
                value={
                  cluster.spec?.backupToS3Enabled ? 'Enabled' : 'Disabled' || ''
                }
              />
              <DetailItem
                title="Kloudlite Release"
                value={cluster.spec?.kloudliteRelease || ''}
              />
              <DetailItem
                title="Public DNS Host"
                value={cluster.spec?.publicDNSHost || ''}
              />
              <DetailItem
                title="Taint Master Nodes"
                value={cluster.spec?.taintMasterNodes ? 'true' : 'false' || ''}
              />
              <DetailItem
                title="Cloud provider"
                value={renderCloudProvider({
                  cloudprovider: cluster.spec?.cloudProvider || 'unknown',
                })}
              />
              {providerInfo()}
            </div>
          </div>
        </Box>
      </div>
    </Wrapper>
  );
};
export default ClusterInfo;
