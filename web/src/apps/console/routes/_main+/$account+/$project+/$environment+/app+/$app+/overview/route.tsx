import { useOutletContext } from '@remix-run/react';
import axios from 'axios';
import Chart from '~/console/components/charts/charts-client';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useState } from 'react';
import { dayjs } from '~/components/molecule/dayjs';
import { parseValue } from '~/console/page-components/util';
import { ApexOptions } from 'apexcharts';
import { IAppContext } from '../route';

const Overview = () => {
  const { app, project } = useOutletContext<IAppContext>();
  const [cpuData, setCpuData] = useState<number[]>([]);
  const [memoryData, setMemoryData] = useState<number[]>([]);

  useDebounce(
    () => {
      (async () => {
        try {
          const resp = await axios({
            url: `https://observe.dev.kloudlite.io/observability/metrics/cpu?cluster_name=${project.clusterName}&tracking_id=${app.id}`,
            method: 'GET',
            withCredentials: true,
          });

          setCpuData(resp?.data?.data?.result[0]?.values || []);
        } catch (err) {
          console.error(err);
        }
      })();
      (async () => {
        try {
          const resp = await axios({
            url: `https://observe.dev.kloudlite.io/observability/metrics/memory?cluster_name=${project.clusterName}&tracking_id=${app.id}`,
            method: 'GET',
            withCredentials: true,
          });

          setMemoryData(resp?.data?.data?.result[0]?.values || []);
        } catch (err) {
          console.error(err);
        }
      })();
    },
    1000,
    []
  );

  const chartOptions: ApexOptions = {
    chart: {
      type: 'area',
      zoom: {
        enabled: false,
      },
      toolbar: {
        show: false,
      },
      redrawOnWindowResize: true,
    },
    dataLabels: {
      enabled: false,
    },
    stroke: {
      curve: 'smooth',
    },
  };

  return (
    <div className="flex gap-6xl items-center h-[30rem] my-6xl">
      <div className="flex-1 h-full">
        <Chart
          title="CPU Usage"
          options={{
            ...chartOptions,
            series: [
              {
                color: '#1D4ED8',
                name: 'CPU',
                data: cpuData,
              },
            ],
            tooltip: {
              x: {
                formatter: (val) => dayjs(val * 1000).format('dd/MM/yy HH:mm'),
              },
              y: {
                formatter(val) {
                  return `${val.toFixed(2)} m`;
                },
              },
            },
            yaxis: {
              min: 0,
              max: parseValue(app.spec.containers[0].resourceCpu?.max, 0),

              floating: false,
              labels: {
                formatter: (val) => `${val} m`,
              },
            },
            xaxis: {
              type: 'datetime',
              labels: {
                formatter(_, timestamp) {
                  return dayjs((timestamp || 0) * 1000).format('hh:mm A');
                },
              },
            },
          }}
        />
      </div>
      <div className="flex-1 h-full">
        <Chart
          title="Memory Usage"
          options={{
            ...chartOptions,
            series: [
              {
                color: '#1D4ED8',
                name: 'Memory',
                data: memoryData,
              },
            ],
            yaxis: {
              min: 0,
              max: parseValue(app.spec.containers[0].resourceMemory?.max, 0),

              floating: false,
              labels: {
                formatter: (val) => `${val} MB`,
              },
            },
            tooltip: {
              x: {
                formatter: (val) => dayjs(val * 1000).format('dd/MM/yy HH:mm'),
              },
              y: {
                formatter(val) {
                  return `${val.toFixed(2)} MB`;
                },
              },
            },
            xaxis: {
              type: 'datetime',
              labels: {
                formatter(_, timestamp) {
                  return dayjs((timestamp || 0) * 1000).format('hh:mm A');
                },
              },
            },
          }}
        />
      </div>
    </div>
  );
};

export default Overview;
