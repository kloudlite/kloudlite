import { Outlet, useOutletContext } from '@remix-run/react';
import { IClusterContext } from '../_layout';

const ContainerRegistry = () => {
  const rootContext = useOutletContext<IClusterContext>();
  return <Outlet context={{ ...rootContext }} />;
};

export default ContainerRegistry;
