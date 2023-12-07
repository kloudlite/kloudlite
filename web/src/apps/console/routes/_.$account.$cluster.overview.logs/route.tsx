import { Suspense } from 'react';
import HighlightJsLog from '~/console/components/logger';

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
    <div className="p-3xl hljs rounded">
      <HighlightJsLog
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
        title={
          <div className="flex flex-col gap-lg">
            <div className="headingXl">Logs</div>
          </div>
        }
        dark
        websocket
        height="60vh"
        width="100%"
        url={getUrl(selectOptions[3].from())}
        selectableLines
      />
    </div>
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
