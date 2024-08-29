import { defer } from '@remix-run/node';
import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import Breadcrum from '~/console/components/breadcrum';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import {
  BackingServices,
  ChevronRight,
  GearSix,
} from '~/console/components/icons';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { IImagePullSecret } from '~/console/server/gql/queries/image-pull-secrets-queries';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseName } from '~/console/server/r-utils/common';
import { ensureAccountSet } from '~/console/server/utils/auth-utils';
import { BreadcrumSlash, tabIconSize } from '~/console/utils/commons';
import logger from '~/lib/client/helpers/log';
import { IRemixCtx } from '~/lib/types/common';
import { IAccountContext } from '../../../_layout';

const ImagePullSecretsTabs = () => {
  const { account, imagepullsecret } = useParams();
  const iconSize = tabIconSize;
  return (
    <CommonTabs
      baseurl={`/${account}/settings/ips/${imagepullsecret}`}
      backButton={{
        to: `/${account}/managed-services`,
        label: 'Integrated Services',
      }}
      tabs={[
        {
          label: (
            <span className="flex flex-row items-center gap-lg">
              <BackingServices size={tabIconSize} />
              Images
            </span>
          ),
          to: '/images',
          value: '/images',
        },
        // {
        //   label: 'Logs & Metrics',
        //   to: '/logs-n-metrics',
        //   value: '/logs-n-metrics',
        // },
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
    />
  );
};

const LocalBreadcrum = ({ data }: { data: IImagePullSecret }) => {
  const { displayName } = data;
  const { account } = useParams();
  return (
    <div className="flex flex-row items-center">
      <BreadcrumSlash />
      <span className="mx-md" />
      <Breadcrum.Button
        to={`/${account}/settings/image-pull-secrets`}
        linkComponent={Link}
        content={
          <div className="flex flex-row gap-md items-center">
            Image Pull Secrets <ChevronRight size={14} />{' '}
          </div>
        }
      />
      <Breadcrum.Button
        to={`/${account}/settings/${parseName(data)}/ips/images`}
        linkComponent={Link}
        content={<span>{displayName}</span>}
      />
    </div>
  );
};

export const handle = ({
  promise: { imagePullSecret, error },
}: {
  promise: any;
}) => {
  if (error) {
    return {};
  }

  return {
    navbar: <ImagePullSecretsTabs />,
    breadcrum: () => <LocalBreadcrum data={imagePullSecret} />,
  };
};

export interface IImagePullSecretContext extends IAccountContext {
  imagepullsecret: IImagePullSecret;
}

const IPSOutlet = ({
  imagePullSecret: OImagePullSecret,
}: {
  imagePullSecret: IImagePullSecret;
}) => {
  const rootContext = useOutletContext<IImagePullSecretContext>();

  return (
    <Outlet context={{ ...rootContext, imagePullSecret: OImagePullSecret }} />
  );
};

export const loader = async (ctx: IRemixCtx) => {
  const promise = pWrapper(async () => {
    ensureAccountSet(ctx);
    const { imagepullsecret } = ctx.params;
    try {
      const { data, errors } = await GQLServerHandler(
        ctx.request
      ).getImagePullSecret({
        name: imagepullsecret,
      });
      if (errors) {
        throw errors[0];
      }

      return {
        imagePullSecret: data,
      };
    } catch (err) {
      logger.log(err);

      return {
        imagePullSecret: {} as IImagePullSecret,
        redirect: `../image-pull-secrets`,
      };
    }
  });
  return defer({ promise: await promise });
};

const ImagePullSecret = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp
      // skeletonData={{
      //   managedService: fake.ConsoleListClusterMSvsQuery
      //     .infra_listClusterManagedServices as any,
      // }}
      data={promise}
    >
      {({ imagePullSecret }) => {
        return <IPSOutlet imagePullSecret={imagePullSecret} />;
      }}
    </LoadingComp>
  );
};

export default ImagePullSecret;
