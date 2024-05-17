import { redirect } from '@remix-run/node';
import {
  Link,
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { SubNavDataProvider } from '~/root/lib/client/hooks/use-create-subnav-action';
import { IRemixCtx } from '~/root/lib/types/common';
import { CommonTabs } from '~/iotconsole/components/common-navbar-tabs';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/iotconsole/server/utils/auth-utils';
import { GQLServerHandler } from '~/iotconsole/server/gql/saved-queries';
import { GearSix, VirtualMachine } from '~/iotconsole/components/icons';
import { ExtractNodeType } from '~/iotconsole/server/r-utils/common';
import LogoWrapper from '~/iotconsole/components/logo-wrapper';
import { BrandLogo } from '~/components/branding/brand-logo';
import { BreadcrumSlash, tabIconSize } from '~/iotconsole/utils/commons';
import { Button } from '~/components/atoms/button';
import {
  IProject,
  IProjects,
} from '~/iotconsole/server/gql/queries/iot-project-queries';
import { IAccountContext } from '../_layout';

export interface IProjectContext extends IAccountContext {
  project: IProject;
}
const iconSize = tabIconSize;
const tabs = [
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <VirtualMachine size={iconSize} />
        Blueprints
      </span>
    ),
    to: '/deviceblueprints',
    value: '/deviceblueprints',
  },
  {
    label: (
      <span className="flex flex-row items-center gap-lg">
        <VirtualMachine size={iconSize} />
        Deployments
      </span>
    ),
    to: '/deployments',
    value: '/deployments',
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
  const rootContext = useOutletContext<IAccountContext>();
  const { project } = useLoaderData();
  return (
    <SubNavDataProvider>
      <Outlet context={{ ...rootContext }} />
    </SubNavDataProvider>
  );
};

const CurrentBreadcrum = ({
  
}: {
  project: ExtractNodeType<IProjects>;
}) => {
  const params = useParams();

  const { account } = params;

  return (
    <>
      <BreadcrumSlash />
      <span className="mx-md" />
      <Button
        content={project.displayName}
        size="sm"
        variant="plain"
        linkComponent={Link}
        to={`/${account}/${project.name}`}
      />
    </>
  );
};

const Tabs = () => {
  const { account } = useParams();

  return <CommonTabs baseurl={`/${account}/${project}`} tabs={tabs} />;
};

const Logo = () => {
  const { account } = useParams();
  return (
    <LogoWrapper to={`/${account}/environments`}>
      <BrandLogo />
    </LogoWrapper>
  );
};

export const handle = ({
  
}: {
  project: ExtractNodeType<IProjects>;
}) => {
  return {
    navbar: <Tabs />,
    breadcrum: () => <CurrentBreadcrum project={project} />,
    logo: <Logo />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { account } = ctx.params;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getIotProject({
      name: 
    });

    if (errors) {
      throw errors[0];
    }

    return {
      project: data || {},
    };
  } catch (err) {
    // logger.error(err);
    return redirect(`/${account}/environments`);
  }
};

export default Project;
