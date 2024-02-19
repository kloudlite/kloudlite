import { useEffect } from 'react';
import { useSocketLogs } from '~/root/lib/client/helpers/socket/useSockLogs';

const App = () => {
  const resp = useSocketLogs({
    account: 'ab-641330',
    cluster: 'ab-cluster-3',
    trackingId: 'app-3ez2fpr-3oc8gqjib-ii5-pbat6d',
  });

  useEffect(() => {
    console.log(resp);
  }, [resp]);

  return <div>Logs</div>;
};

export default App;
