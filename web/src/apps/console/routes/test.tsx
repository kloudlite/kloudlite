import { useLoaderData } from '@remix-run/react';
import SocketProvider, {
  useSubscribe,
  // useSubscribe,
  useWatch,
} from '~/root/lib/client/helpers/socket-context';

export const loader = () => {
  return {
    data: Math.random(),
  };
};

const App = () => {
  useSubscribe(
    ['account:newteam.cluster'],
    () => {
      console.log('hi');
    },
    []
  );

  // useWatch('account:newteam.cluster');
  // res-updates.account.acc-ruwibp-pf5jvcsew2rnl54kriv59.cluster.*
  // res-updates.account.accid.project.projid.env.envid.app.*

  const { data } = useLoaderData();

  return <div>{data}</div>;
};

export default () => {
  return (
    <SocketProvider>
      <App />
    </SocketProvider>
  );
};
