import { Outlet, useOutletContext } from '@remix-run/react';
import { IEnvironmentContext } from '../_layout';

const WorkspaceSettings = () => {
  const rootContext = useOutletContext<IEnvironmentContext>();

  return <Outlet context={{ ...rootContext }} />;
};

export default WorkspaceSettings;
