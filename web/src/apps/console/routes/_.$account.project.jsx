import { Outlet, useOutletContext } from '@remix-run/react';

const Projects = () => {
  const rootContext = useOutletContext();
  return <Outlet context={rootContext} />;
};

export default Projects;
