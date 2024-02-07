/* eslint-disable react/no-unused-prop-types */
import ApexCharts from 'apexcharts';
import { type ApexOptions } from 'apexcharts';
import { useEffect, useRef } from 'react';

const ChartServer = ({ options }: { options: ApexOptions }) => {
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
