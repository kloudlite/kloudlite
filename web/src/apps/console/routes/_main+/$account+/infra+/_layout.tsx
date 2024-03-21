import { Outlet, useOutletContext } from '@remix-run/react';

const Infra = () => {
  const rootContext = useOutletContext();
  return <Outlet context={rootContext} />;
};

export default Infra;
