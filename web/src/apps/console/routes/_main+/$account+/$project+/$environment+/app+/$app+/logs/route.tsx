import { useOutletContext } from '@remix-run/react';
import LogComp from '~/console/components/logger';
import { IAppContext } from '../route';

const Overview = () => {
  const { app, project, account } = useOutletContext<IAppContext>();
  return (
    <div className="my-6xl flex-1">
      <LogComp
        {...{
          dark: true,
          width: '100%',
          height: 'calc(100vh - 12rem)',
          title: 'Logs',
          websocket: {
            account: account.metadata?.name,
            cluster: project.clusterName,
            trackingId: app.id,
          },
        }}
      />
    </div>
  );
};

export default Overview;
