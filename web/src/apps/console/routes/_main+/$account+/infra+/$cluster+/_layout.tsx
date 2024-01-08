import { redirect } from '@remix-run/node';
import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import withContext from '~/root/lib/app-setup/with-contxt';
import { IExtRemixCtx } from '~/root/lib/types/common';
import { BrandLogo } from '~/components/branding/brand-logo';
import {
  ChevronRight,
  Cpu,
  Crosshair,
  Database,
  GearSix,
  HouseLine,
  WireGuardlogo,
} from '@jengaicons/react';
import {
  ICluster,
  IClusters,
} from '~/console/server/gql/queries/cluster-queries';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import Breadcrum from '~/console/components/breadcrum';
import LogoWrapper from '~/console/components/logo-wrapper';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { IAccountContext } from '../../_layout';

export interface IClusterContext extends IAccountContext {
  cluster: ICluster;
}

const Cluster = () => {
  const rootContext = useOutletContext<IAccountContext>();
  const { cluster } = useLoaderData();
  return <Outlet context={{ ...rootContext, cluster }} />;
};

const ClusterTabs = () => {
  const { account, cluster } = useParams();
  const iconSize = 16;
  return (
    <CommonTabs
      tabs={[
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <HouseLine size={iconSize} />
              Overview
            </span>
          ),
          to: '/overview',
          value: '/overview',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <Cpu size={iconSize} />
              Nodepools
            </span>
          ),
          to: '/nodepools',
          value: '/nodepools',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <Database size={iconSize} />
              Storage
            </span>
          ),
          to: '/storage',
          value: '/storage',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <WireGuardlogo size={iconSize} />
              VPN
            </span>
          ),
          to: '/network/vpn',
          value: '/network',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <Crosshair size={iconSize} />
              Helm charts
            </span>
          ),
          to: '/helm-charts',
          value: '/helm-charts',
        },
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <GearSix size={iconSize} />
              Settings
            </span>
          ),
          to: '/settings/general',
          value: '/settings',
        },
      ]}
      baseurl={`/${account}/infra/${cluster}`}
    />
  );
};

const NetworkBreadcrum = ({
  cluster,
}: {
  cluster: ExtractNodeType<IClusters>;
}) => {
  const { displayName } = cluster;
  const { account } = useParams();
  return (
    <div className="flex flex-row items-center">
      <Breadcrum.Button
        to={`/${account}/infra/clusters`}
        LinkComponent={Link}
        content={
          <div className="flex flex-row gap-md items-center">
            <ChevronRight size={14} /> Clusters <ChevronRight size={14} />{' '}
          </div>
        }
      />
      <Breadcrum.Button
        to={`/${account}/infra/${parseName(cluster)}/overview/info`}
        LinkComponent={Link}
        content={<span>{displayName}</span>}
      />
    </div>
  );
};

const Logo = () => {
  const { account } = useParams();
  return (
    <LogoWrapper to={`/${account}`}>
      <BrandLogo detailed={false} />
    </LogoWrapper>
  );
};

export const handle = ({
  cluster,
}: {
  cluster: ExtractNodeType<IClusters>;
}) => {
  return {
    navbar: <ClusterTabs />,
    breadcrum: () => <NetworkBreadcrum cluster={cluster} />,
    logo: <Logo />,
    noLayout: true,
  };
};

export const loader = async (ctx: IExtRemixCtx) => {
  const { account, cluster } = ctx.params;
  ensureAccountSet(ctx);
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getCluster({
      name: cluster,
    });
    if (errors) {
      throw errors[0];
    }
    ensureClusterSet(ctx);
    return withContext(ctx, {
      cluster: data,
    });
  } catch (err) {
    return redirect(`/${account}/infra/clusters`);
  }
};

export default Cluster;
