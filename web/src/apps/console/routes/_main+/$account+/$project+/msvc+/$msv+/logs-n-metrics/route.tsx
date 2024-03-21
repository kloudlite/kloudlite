import { useOutletContext } from '@remix-run/react';
import axios from 'axios';
import Chart from '~/console/components/charts/charts-client';
import useDebounce from '~/lib/client/hooks/use-debounce';
import { useState } from 'react';
import { dayjs } from '~/components/molecule/dayjs';
import { parseValue } from '~/console/page-components/util';
import { ApexOptions } from 'apexcharts';
import { useDataState } from '~/console/page-components/common-state';
import { observeUrl } from '~/lib/configs/base-url.cjs';
import LogComp from '~/lib/client/components/logger';
import LogAction from '~/console/page-components/log-action';
import { IProjectManagedServiceContext } from '~/console/routes/_main+/$account+/$project+/msvc+/$msv+/_layout';
import { parseName } from '~/console/server/r-utils/common';
import { generatePlainColor } from '~/root/lib/utils/color-generator';

const LogsAndMetrics = () => {
  const { account, project, managedService } =
    useOutletContext<IProjectManagedServiceContext>();
  type tData = {
    metric: {
      exported_pod: string;
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
            url: `${observeUrl}/observability/metrics/cpu?cluster_name=${project.clusterName}&tracking_id=${managedService.id}`,
            method: 'GET',
            withCredentials: true,
          });

          setData((prev) => ({
            ...prev,
            cpu: resp?.data?.data?.result || [],
          }));

          // setCpuData(resp?.data?.data?.result[0]?.values || []);
        } catch (err) {
          console.error('error1', err);
        }
      })();
      (async () => {
        try {
          const resp = await axios({
            url: `${observeUrl}/observability/metrics/memory?cluster_name=${project.clusterName}&tracking_id=${managedService.id}`,
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
                  name: d.metric.exported_pod,
                  color: generatePlainColor(d.metric.exported_pod),
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
              managedService.spec?.msvcSpec.serviceTemplate.spec?.resources
                ?.cpu || {},

              'cpu'
            ),

            yaxis: {
              min: 0,
              max:
                parseValue(
                  managedService.spec?.msvcSpec.serviceTemplate.spec?.resources
                    ?.cpu?.max,
                  0
                ) * 1.1,

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
                  name: d.metric.exported_pod,
                  color: generatePlainColor(d.metric.exported_pod),
                  data: d.values.map((v) => {
                    return [v[0], parseFloat(v[1])];
                  }),
                };
              }),
            ],

            annotations: getAnnotations(
              managedService.spec?.msvcSpec.serviceTemplate.spec?.resources
                ?.cpu || {},
              'memory'
            ),

            yaxis: {
              min: 0,
              max:
                parseValue(
                  managedService.spec?.msvcSpec.serviceTemplate.spec?.resources
                    ?.cpu?.max,
                  0
                ) * 1.1,

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
            dark: true,
            width: '100%',
            height: '70vh',
            title: 'Logs',
            actionComponent: <LogAction />,
            websocket: {
              account: parseName(account),
              cluster: project.clusterName || '',
              trackingId: managedService.id,
              recordVersion: managedService.recordVersion,
            },
          }}
        />
      </div>
    </div>
  );
};

export default LogsAndMetrics;
