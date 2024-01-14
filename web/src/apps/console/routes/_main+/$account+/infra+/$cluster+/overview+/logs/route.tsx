import { Suspense } from 'react';
import { Box } from '~/console/components/common-console-components';
import LogComp from '~/console/components/logger';

const Log = () => {
  const getTime = () => {
    return Math.floor(new Date().getTime() / 1000);
  };

  const selectOptions = [
    {
      label: 'Last 12 hours',
      value: '1',
      from: () => getTime() - 43200,
    },
    {
      label: 'Last 24 hours',
      value: '2',
      from: () => getTime() - 86400,
    },
    {
      label: 'Last 7 days',
      value: '3',
      from: () => getTime() - 604800,
    },
    {
      label: 'Last 30 days',
      value: '3',
      from: () => getTime() - 2592000,
    },
  ];

  // const [from, setFrom] = useState(selectOptions[1].from());
  // const [selected, setSelected] = useState('1');

  const getUrl = (f: number) => {
    return `wss://observability.dev.kloudlite.io/observability/logs/cluster-job?start_time=${f}&end_time=${getTime()}`;
  };

  // const [url, setUrl] = useState(getUrl(from));

  return (
    <Box title="Cluster Logs">
      <div className=" hljs rounded">
        <LogComp
          // actionComponent={
          //   <Select
          //     size="md"
          //     options={async () => selectOptions}
          //     value={selectOptions[parseValue(selected, 1)]}
          //     onChange={(e) => {
          //       setSelected(e.value);
          //     }}
          //   />
          // }
          // title={
          //   <div className="flex flex-col gap-lg">
          //     <div className="">Cluster logs</div>
          //   </div>
          // }
          dark
          websocket={{
            account: '',
            trackingId: '',
            cluster: '',
          }}
          height="60vh"
          width="100%"
          url={getUrl(selectOptions[3].from())}
          selectableLines
        />
      </div>
    </Box>
  );
};

const ClusterLogs = () => {
  return (
    <div>
      <Suspense>
        <Log />
      </Suspense>
    </div>
  );
};

export default ClusterLogs;
