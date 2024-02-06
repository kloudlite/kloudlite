/* eslint-disable react/no-unused-prop-types */
import ApexCharts from 'apexcharts';
import { useEffect, useRef } from 'react';

export interface ChartProps {
  height?: string;
  width?: string;
  options: {
    chart: {
      height: number;
      type: string;
    };
    series: {
      name: string;
      data: number[];
    }[];
    dataLabels: {
      enabled: boolean;
    };
    stroke: {
      curve: string;
    };
    tooltip: {
      x: {
        format: string;
      };
    };
    xaxis: {
      type: 'category' | 'numeric' | 'logarithmic' | 'datetime';
      categories: (number | string)[];
    };
  };
}

const ChartServer = ({ options }: ChartProps) => {
  const ref = useRef(null);

  useEffect(() => {
    if (!ref.current) {
      return () => {};
    }

    const chart = new ApexCharts(ref.current, options);
    chart.render();

    return () => {
      chart.destroy();
    };
  }, [ref.current, options]);

  return <div ref={ref} />;
};

export default ChartServer;
