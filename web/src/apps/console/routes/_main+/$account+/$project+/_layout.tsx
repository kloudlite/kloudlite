import { redirect } from '@remix-run/node';
import {
  Link,
  Outlet,
  useLoaderData,
  useNavigate,
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
import {
  Database,
  GearSix,
  VirtualMachine,
  InfraAsCode,
  Container as ContainerIcon,
  Project as ProjectIcon,
} from '@jengaicons/react';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import MenuSelect from '~/console/components/menu-select';
import LogoWrapper from '~/console/components/logo-wrapper';
import { BrandLogo } from '~/components/branding/brand-logo';
import {
  BreadcrumButtonContent,
  BreadcrumSlash,
} from '~/console/utils/commons';
import { IClusterContext } from '../infra+/$cluster+/_layout';

export interface IProjectContext extends IClusterContext {
  project: IProject;
}

const Project = () => {
  const rootContext = useOutletContext<IClusterContext>();
  const { project } = useLoaderData();
  return (
    <SubNavDataProvider>
      <Outlet context={{ ...rootContext, project }} />
    </SubNavDataProvider>
  );
};

const ProjectsDropdown = () => {
  const navigate = useNavigate();
  const { account } = useParams();
  const iconSize = 14;

  const menuItems = [
    {
      label: (
        <span className="flex flex-row items-center gap-lg">
          <ProjectIcon size={iconSize} />
          Projects
        </span>
      ),
      value: `/${account}/projects`,
    },
    {
      label: (
        <span className="flex flex-row items-center gap-lg">
          <InfraAsCode size={iconSize} />
          Infrastructure
        </span>
      ),
      value: `/${account}/infra`,
    },
    {
      label: (
        <span className="flex flex-row items-center gap-lg">
          <ContainerIcon size={iconSize} />
          Packages
        </span>
      ),
      value: `/${account}/packages`,
    },
  ];
  return (
    <MenuSelect
      items={menuItems}
      value={`/${account}/projects`}
      onClick={(value) => navigate(value)}
      trigger={
        <Breadcrum.Button
          content={<BreadcrumButtonContent content="Projects" />}
        />
      }
    />
  );
};

const LocalBreadcrum = ({
  project,
}: {
  project: ExtractNodeType<IProjects>;
}) => {
  const iconSize = 14;
  const { account, cluster } = useParams();
  return (
    <div className="flex flex-row items-center">
      {/* <ProjectsDropdown project={project} /> */}
      <BreadcrumSlash />
      <Breadcrum.Button
        to={`/${account}/${cluster}/${parseName(project)}/environments`}
        LinkComponent={Link}
        content={
          <div className="flex flex-row items-center">
            <BreadcrumButtonContent content={project.displayName} />
            {/* <span className="capitalize"></span> */}
          </div>
        }
      />
    </div>
  );
};

const ProjectTabs = () => {
  const { account, project } = useParams();
  const iconSize = 16;
  return (
    <CommonTabs
      baseurl={`/${account}/${project}`}
      tabs={[
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
      ]}
    />
  );
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
    navbar: <ProjectTabs />,
    breadcrum: () => <LocalBreadcrum project={project} />,
    logo: <Logo />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { account, project, cluster } = ctx.params;
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
    return redirect(`/${account}/${cluster}/projects`);
  }
};

export default Project;
