import { Outlet, useOutletContext } from '@remix-run/react';

const Builds = () => {
  const rootContext = useOutletContext();
  return <Outlet context={rootContext} />;
};

export default Builds;
