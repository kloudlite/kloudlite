import { Outlet, useOutletContext } from '@remix-run/react';

const ProjectConfig = () => {
  const [subNavAction, setSubNavAction] = useOutletContext();
  return <Outlet context={[subNavAction, setSubNavAction]} />;
};

export default ProjectConfig;
