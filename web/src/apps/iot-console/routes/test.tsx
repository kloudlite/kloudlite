import { useLoaderData } from '@remix-run/react';

export const loader = () => {
  return {
    data: Math.random(),
  };
};

const App = () => {
  // useWatch('account:newteam.cluster');
  // res-updates.account.acc-ruwibp-pf5jvcsew2rnl54kriv59.cluster.*
  // res-updates.account.accid.project.projid.env.envid.app.*

  const { data } = useLoaderData();

  return <div>{data}</div>;
};

export default () => {
  return <App />;
};
