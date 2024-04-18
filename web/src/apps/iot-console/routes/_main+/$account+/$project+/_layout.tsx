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
      <Outlet context={{ ...rootContext, project }} />
    </SubNavDataProvider>
  );
};

const CurrentBreadcrum = ({
  project,
}: {
  project: ExtractNodeType<IProjects>;
}) => {
  const params = useParams();

  // const api = useIotConsoleApi();
  // const [search, setSearch] = useState('');
  // const [searchText, setSearchText] = useState('');

  const { account } = params;

  // const { data: projects, isLoading } = useCustomSwr(
  //   () => `/projects/${searchText}`,
  //   async () =>
  //     api.listProjects({
  //       search: {
  //         text: {
  //           matchType: 'regex',
  //           regex: searchText,
  //         },
  //       },
  //     })
  // );

  // useDebounce(
  //   async () => {
  //     ensureAccountClientSide(params);
  //     setSearchText(search);
  //   },
  //   300,
  //   [search]
  // );

  // const [open, setOpen] = useState(false);
  // const buttonRef = useRef<HTMLButtonElement>(null);
  // const [isMouseOver, setIsMouseOver] = useState<boolean>(false);

  return (
    <>
      <BreadcrumSlash />
      <span className="mx-md" />
      <Button
        content={project.displayName}
        size="sm"
        variant="plain"
        LinkComponent={Link}
        to={`/${account}/${project.name}`}
      />
      {/* <OptionList.Root open={open} onOpenChange={setOpen} modal={false}> */}
      {/*   <OptionList.Trigger> */}
      {/*     <button */}
      {/*       ref={buttonRef} */}
      {/*       aria-label="accounts" */}
      {/*       className={cn( */}
      {/*         'outline-none rounded py-lg px-md mx-md bg-surface-basic-hovered', */}
      {/*         open || isMouseOver ? 'bg-surface-basic-pressed' : '' */}
      {/*       )} */}
      {/*       onMouseOver={() => { */}
      {/*         setIsMouseOver(true); */}
      {/*       }} */}
      {/*       onMouseOut={() => { */}
      {/*         setIsMouseOver(false); */}
      {/*       }} */}
      {/*       onFocus={() => { */}
      {/*         // */}
      {/*       }} */}
      {/*       onBlur={() => { */}
      {/*         // */}
      {/*       }} */}
      {/*     > */}
      {/*       <div className="flex flex-row items-center gap-md"> */}
      {/*         <ChevronUpDown size={16} /> */}
      {/*       </div> */}
      {/*     </button> */}
      {/*   </OptionList.Trigger> */}
      {/*   <OptionList.Content className="!pt-0 !pb-md" align="end"> */}
      {/*     <div className="p-[3px] pb-0"> */}
      {/*       <OptionList.TextInput */}
      {/*         value={search} */}
      {/*         onChange={(e) => setSearch(e.target.value)} */}
      {/*         prefixIcon={<Search />} */}
      {/*         focusRing={false} */}
      {/*         placeholder="Search projects" */}
      {/*         compact */}
      {/*         className="border-0 rounded-none" */}
      {/*       /> */}
      {/*     </div> */}
      {/*     <OptionList.Separator /> */}

      {/*     {!isLoading && */}
      {/*       (parseNodes(projects) || [])?.map((item) => { */}
      {/*         return ( */}
      {/*           <OptionList.Link */}
      {/*             key={parseName(item)} */}
      {/*             LinkComponent={Link} */}
      {/*             to={`/${account}/${parseName(item)}/environments`} */}
      {/*             className={cn( */}
      {/*               'flex flex-row items-center justify-between', */}
      {/*               parseName(item) === parseName(project) */}
      {/*                 ? 'bg-surface-basic-pressed hover:!bg-surface-basic-pressed' */}
      {/*                 : '' */}
      {/*             )} */}
      {/*           > */}
      {/*             <span>{item.displayName}</span> */}
      {/*             {parseName(item) === parseName(project) && ( */}
      {/*               <span> */}
      {/*                 <Check size={16} /> */}
      {/*               </span> */}
      {/*             )} */}
      {/*           </OptionList.Link> */}
      {/*         ); */}
      {/*       })} */}

      {/*     {parseNodes(projects).length === 0 && !isLoading && ( */}
      {/*       <div className="flex flex-col gap-lg max-w-[198px] px-xl py-lg"> */}
      {/*         <div className="bodyLg-medium text-text-default"> */}
      {/*           No projects found */}
      {/*         </div> */}
      {/*         <div className="bodyMd text-text-soft"> */}
      {/*           Your search for "{search}" did not match and projects. */}
      {/*         </div> */}
      {/*       </div> */}
      {/*     )} */}

      {/*     {isLoading && parseNodes(projects).length === 0 && ( */}
      {/*       <div className="min-h-7xl" /> */}
      {/*     )} */}

      {/*     <OptionList.Separator /> */}
      {/*     <OptionList.Link */}
      {/*       LinkComponent={Link} */}
      {/*       to={`/${account}/new-project`} */}
      {/*       className="text-text-primary" */}
      {/*     > */}
      {/*       <Plus size={16} /> <span>Create project</span> */}
      {/*     </OptionList.Link> */}
      {/*   </OptionList.Content> */}
      {/* </OptionList.Root> */}
    </>
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
    breadcrum: () => <CurrentBreadcrum project={project} />,
    logo: <Logo />,
  };
};

export const loader = async (ctx: IRemixCtx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { account, project } = ctx.params;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getIotProject({
      name: project,
    });

    if (errors) {
      throw errors[0];
    }

    return {
      project: data || {},
    };
  } catch (err) {
    // logger.error(err);
    return redirect(`/${account}/projects`);
  }
};

export default Project;
