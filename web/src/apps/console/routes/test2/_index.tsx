import LogComp from '~/console/components/logger';

const App = () => {
  return (
    <div className="flex items-center justify-center h-screen w-screen">
      <LogComp
        {...{
          dark: true,
          width: '80vw',
          height: '80vh',
          title: 'Logs',
          websocket: {
            account: 'kloudlite-dev',
            cluster: 'sample-cluster',
            trackingId: 'app-k-zmtg0km7epjj-fq89uvao14-3-l',
          },
        }}
      />
    </div>
  );
};

const Logs = () => {
  return <App />;
};

export default Logs;
