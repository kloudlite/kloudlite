import { Outlet } from '@remix-run/react';

const Project = () => {
  return <Outlet />;
};

export default Project;

export const handle = {
  navbar: [
    {
      label: 'Apps',
      href: '/apps',
      key: 'apps',
      value: '/apps',
    },
    {
      label: 'Routers',
      href: '/routers',
      key: 'routers',
      value: '/routers',
    },
    {
      label: 'Config & Secrets',
      href: '/config-and-secrets',
      key: 'config-and-secrets',
      value: '/config-and-secrets',
    },
    {
      label: 'Backing services',
      href: '/backing-services',
      key: 'backing-services',
      value: '/backing-services',
    },
  ],
};
export const loader = async (ctx) => {
  return { baseurl: `/${ctx.params.account}/projects/${ctx.params.project}` };
};
