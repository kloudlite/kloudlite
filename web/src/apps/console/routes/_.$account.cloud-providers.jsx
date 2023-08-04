import { Outlet, useOutletContext } from '@remix-run/react';

const CloudProviders = () => {
  const rootContext = useOutletContext();
  return <Outlet context={rootContext} />;
};

export default CloudProviders;
