import { redirect } from '@remix-run/node';
import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import logger from '~/root/lib/client/helpers/log';
import { SubNavDataProvider } from '~/root/lib/client/hooks/use-create-subnav-action';
import { IRemixCtx } from '~/root/lib/types/common';
import {
  IProject,
  IProjects,
} from '~/console/server/gql/queries/project-queries';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import Breadcrum from '~/console/components/breadcrum';
import { Database, GearSix, VirtualMachine } from '@jengaicons/react';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import LogoWrapper from '~/console/components/logo-wrapper';
import { BrandLogo } from '~/components/branding/brand-logo';
import {
  BreadcrumButtonContent,
  BreadcrumSlash,
  tabIconSize,
} from '~/console/utils/commons';
import { useActivePath } from '~/root/lib/client/hooks/use-active-path';
import { cn } from '~/components/utils';
import { IClusterContext } from '../infra+/$cluster+/_layout';

export interface IProjectContext extends IClusterContext {
  project: IProject;
}
const iconSize = tabIconSize;
const tabs = [
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <VirtualMachine size={iconSize} />
        Environments
      </span>
    ),
    to: '/environments',
    value: '/environments',
  },
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <Database size={iconSize} />
        Managed services
      </span>
    ),
    to: '/managed-services',
    value: '/managed-services',
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
];

const Project = () => {
  const rootContext = useOutletContext<IClusterContext>();
  const { project } = useLoaderData();
  return (
    <SubNavDataProvider>
      <Outlet context={{ ...rootContext, project }} />
    </SubNavDataProvider>
  );
};

const LocalBreadcrum = ({
  project,
}: {
  project: ExtractNodeType<IProjects>;
}) => {
  const { account } = useParams();
  const { activePath } = useActivePath({
    parent: `/${account}/${parseName(project)}`,
  });

  return (
    <div className="flex flex-row items-center">
      <BreadcrumSlash />
      <Breadcrum.Button
        to={`/${account}/${parseName(project)}/environments`}
        LinkComponent={Link}
        content={
          <div
            className={cn(
              'flex flex-row items-center',
              tabs.find((tab) => tab.to === activePath) ? 'bodyMd-semibold' : ''
            )}
          >
            <BreadcrumButtonContent content={project.displayName} />
          </div>
        }
      />
    </div>
  );
};

const Tabs = () => {
  const { account, project } = useParams();

  return <CommonTabs baseurl={`/${account}/${project}`} tabs={tabs} />;
};

const Logo = () => {
  const { account } = useParams();
  return (
    <LogoWrapper to={`/${account}/projects`}>
      <BrandLogo />
    </LogoWrapper>
  );
};

export const handle = ({
  project,
}: {
  project: ExtractNodeType<IProjects>;
}) => {
  return {
    navbar: <Tabs />,
    breadcrum: () => <LocalBreadcrum project={project} />,
    logo: <Logo />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { account, project } = ctx.params;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getProject({
      name: project,
    });
    if (errors) {
      throw errors[0];
    }
    return {
      project: data || {},
    };
  } catch (err) {
    logger.log(err);
    return redirect(`/${account}/projects`);
  }
};

export default Project;
