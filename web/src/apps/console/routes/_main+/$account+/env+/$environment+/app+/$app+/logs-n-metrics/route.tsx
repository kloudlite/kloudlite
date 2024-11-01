import { useOutletContext } from '@remix-run/react';
import { ApexOptions } from 'apexcharts';
import axios from 'axios';
import { useState } from 'react';
import { dayjs } from '@kloudlite/design-system/molecule/dayjs';
import Chart from '~/console/components/charts/charts-client';
import { findClusterStatusv3 } from '~/console/hooks/use-cluster-status';
import { useClusterStatusV3 } from '~/console/hooks/use-cluster-status-v3';
import { useDataState } from '~/console/page-components/common-state';
import LogAction from '~/console/page-components/log-action';
import { NoLogsAndMetricsBanner } from '~/console/page-components/no-logs-banner';
import { parseValue } from '~/console/page-components/util';
import { parseName } from '~/console/server/r-utils/common';
import LogComp from '~/lib/client/components/logger';
import useDebounce from '~/lib/client/hooks/use-debounce';
import { observeUrl } from '~/lib/configs/base-url.cjs';
import { generatePlainColor } from '~/root/lib/utils/color-generator';
import { IAppContext } from '../_layout';

const LogsAndMetrics = () => {
  const { app, environment, account } = useOutletContext<IAppContext>();

  const { clustersMap: clusterStatus } = useClusterStatusV3({
    clusterName: environment.clusterName,
  });

  const isClusterOnline = findClusterStatusv3(
    clusterStatus[environment.clusterName]
  );

  type tData = {
    metric: {
      pod_name: string;
    };
    values: [number, string][];
  };

  const [data, setData] = useState<{
    cpu: tData[];
    memory: tData[];
  }>({
    cpu: [],
    memory: [],
  });

  const xAxisFormatter = (val: string, __?: number) => {
    return dayjs((parseValue(val, 0) || 0) * 1000).format('hh:mm A');
    // return '';
  };

  const tooltipXAixsFormatter = (val: number) =>
    dayjs(val * 1000).format('DD/MM/YY hh:mm A');

  const getAnnotations = (
    {
      min = '',
      max = '',
    }: {
      min?: string;
      max?: string;
    },
    resType: 'cpu' | 'memory'
  ) => {
    const tmin = parseValue(min, 0);
    const tmax = parseValue(max, 0);

    const unit = resType === 'cpu' ? 'vCPU' : 'MB';

    const k: ApexOptions['annotations'] = {
      yaxis: [
        {
          y: tmin,
          y2: tmin === tmax ? tmax + 1 : tmax,
          fillColor: '#33f',
          borderColor: '#33f',
          opacity: 0.1,
          strokeDashArray: 0,
          borderWidth: 1,
          label: {
            style: {
              fontFamily: 'Inter',
              fontSize: '14px',
            },
            // textAnchor: 'middle',
            // position: 'center',
            text:
              tmin === tmax
                ? `allocated: ${tmax}${unit}`
                : `min: ${tmin}${unit} | max: ${tmax}${unit}`,
          },
        },
      ],
    };

    return k;
  };

  useDebounce(
    () => {
      (async () => {
        try {
          const resp = await axios({
            url: `${observeUrl}/observability/metrics/cpu?cluster_name=${environment.clusterName}&tracking_id=${app.id}`,
            method: 'GET',
            withCredentials: true,
          });

          setData((prev) => ({
            ...prev,
            cpu: resp?.data?.data?.result || [],
          }));

          // setCpuData(resp?.data?.data?.result[0]?.values || []);
        } catch (err) {
          console.error(err);
        }
      })();
      (async () => {
        try {
          const resp = await axios({
            url: `${observeUrl}/observability/metrics/memory?cluster_name=${environment.clusterName}&tracking_id=${app.id}`,
            method: 'GET',
            withCredentials: true,
          });

          setData((prev) => ({
            ...prev,
            memory: resp?.data?.data?.result || [],
          }));
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
      type: 'line',
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
      width: 2,
      curve: 'smooth',
    },

    xaxis: {
      type: 'datetime',
      labels: {
        show: false,
        formatter: xAxisFormatter,
      },
    },
  };

  const { state } = useDataState<{
    linesVisible: boolean;
    timestampVisible: boolean;
  }>('logs');

  if (environment.clusterName === '') {
    return (
      <NoLogsAndMetricsBanner
        title="Logs and Metrics Unavailable for Template Applications"
        description="Logs and metrics will become available once your app is associated with a cluster."
      />
    );
  }

  if (isClusterOnline === false) {
    return (
      <NoLogsAndMetricsBanner
        title="Logs and Metrics Unavailable for Offline Cluster-Based Applications"
        description="Logs and metrics will be accessible once the cluster is back online."
      />
    );
  }

  return (
    <div className="flex flex-col gap-6xl pt-6xl">
      <div className="gap-6xl items-center flex-col grid sm:grid-cols-2 lg:grid-cols-2">
        <Chart
          title="CPU Usage"
          options={{
            ...chartOptions,
            series: [
              ...data.cpu.map((d) => {
                return {
                  name: d.metric.pod_name,
                  color: generatePlainColor(d.metric.pod_name),
                  data: d.values.map((v) => {
                    return [v[0], parseFloat(v[1])];
                  }),
                };
              }),
            ],
            tooltip: {
              x: {
                formatter: tooltipXAixsFormatter,
              },
              y: {
                formatter(val) {
                  return `${val.toFixed(2)} m`;
                },
              },
            },

            annotations: getAnnotations(
              app.spec.containers[0].resourceCpu || {},
              'cpu'
            ),

            yaxis: {
              min: 0,
              max: parseValue(app.spec.containers[0].resourceCpu?.max, 0) * 1.1,

              floating: false,
              labels: {
                formatter: (val) => {
                  return `${(val / 1000).toFixed(3)} vCPU`;
                },
              },
            },
          }}
        />

        <Chart
          title="Memory Usage"
          options={{
            ...chartOptions,
            series: [
              ...data.memory.map((d) => {
                return {
                  name: d.metric.pod_name,
                  color: generatePlainColor(d.metric.pod_name),
                  data: d.values.map((v) => {
                    return [v[0], parseFloat(v[1])];
                  }),
                };
              }),
            ],

            annotations: getAnnotations(
              app.spec.containers[0].resourceMemory || {},
              'memory'
            ),

            yaxis: {
              min: 0,
              max:
                parseValue(app.spec.containers[0].resourceMemory?.max, 0) * 1.1,

              floating: false,
              labels: {
                formatter: (val) => `${val.toFixed(0)} MB`,
              },
            },
            tooltip: {
              x: {
                formatter: tooltipXAixsFormatter,
              },
              y: {
                formatter(val) {
                  return `${val.toFixed(2)} MB`;
                },
              },
            },
          }}
        />
      </div>

      <div className="flex-1">
        <LogComp
          {...{
            hideLineNumber: !state.linesVisible,
            hideTimestamp: !state.timestampVisible,
            podSelect: true,
            dark: true,
            width: '100%',
            height: '70vh',
            title: 'Logs',
            actionComponent: <LogAction />,
            websocket: {
              account: parseName(account),
              cluster: environment.clusterName || '',
              trackingId: app.id,
              recordVersion: app.recordVersion,
            },
          }}
        />
      </div>
    </div>
  );
};

export default LogsAndMetrics;
