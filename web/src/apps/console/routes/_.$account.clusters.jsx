import { Outlet, useOutletContext } from '@remix-run/react';

const Clusters = () => {
  const rootContext = useOutletContext();
  return <Outlet context={rootContext} />;
};

export default Clusters;
