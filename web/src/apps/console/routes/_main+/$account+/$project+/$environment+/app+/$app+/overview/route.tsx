import { useOutletContext } from '@remix-run/react';
import axios from 'axios';
import { useEffect } from 'react';
import Chart from '~/console/components/charts/charts-client';
import { IAppContext } from '../route';

const Overview = () => {
  const { app, project } = useOutletContext<IAppContext>();
  useEffect(() => {
    (async () => {
      try {
        const resp = await axios({
          url: `https://observe.dev.kloudlite.io/observability/metrics/memory?cluster_name=${project.clusterName}&tracking_id=${app.id}`,
          method: 'GET',
          withCredentials: true,
        });

        console.log(resp);
      } catch (err) {
        console.error(err);
      }
    })();
  }, []);

  return (
    <div className="flex gap-6xl items-center h-[30rem]">
      <div className="flex-1">
        <Chart
          options={{
            series: [
              {
                name: 'series1',
                data: [31, 40, 28, 51, 42, 109, 100],
              },
              {
                name: 'series2',
                data: [11, 32, 45, 32, 34, 52, 41],
              },
            ],
            chart: {
              height: 350,
              type: 'area',
            },
            dataLabels: {
              enabled: false,
            },
            stroke: {
              curve: 'smooth',
            },
            xaxis: {
              type: 'datetime',
              categories: [
                '2018-09-19T00:00:00.000Z',
                '2018-09-19T01:30:00.000Z',
                '2018-09-19T02:30:00.000Z',
                '2018-09-19T03:30:00.000Z',
                '2018-09-19T04:30:00.000Z',
                '2018-09-19T05:30:00.000Z',
                '2018-09-19T06:30:00.000Z',
              ],
            },
            tooltip: {
              x: {
                format: 'dd/MM/yy HH:mm',
              },
            },
          }}
        />
      </div>
      <div className="flex-1">
        <Chart
          options={{
            series: [
              {
                name: 'series1',
                data: [31, 40, 28, 51, 42, 109, 100],
              },
              {
                name: 'series2',
                data: [11, 32, 45, 32, 34, 52, 41],
              },
            ],
            chart: {
              height: 350,
              type: 'area',
            },
            dataLabels: {
              enabled: false,
            },
            stroke: {
              curve: 'smooth',
            },
            xaxis: {
              type: 'datetime',
              categories: [
                '2018-09-19T00:00:00.000Z',
                '2018-09-19T01:30:00.000Z',
                '2018-09-19T02:30:00.000Z',
                '2018-09-19T03:30:00.000Z',
                '2018-09-19T04:30:00.000Z',
                '2018-09-19T05:30:00.000Z',
                '2018-09-19T06:30:00.000Z',
              ],
            },
            tooltip: {
              x: {
                format: 'dd/MM/yy HH:mm',
              },
            },
          }}
        />
      </div>
    </div>
  );
};

export default Overview;
