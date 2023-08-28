import { Outlet, useOutletContext } from '@remix-run/react';

const ProjectConfig = () => {
  const rootContext = useOutletContext();
  return <Outlet context={rootContext} />;
};

export default ProjectConfig;
