import React, { Suspense } from 'react';
import { cn } from '~/components/utils';
import { type ApexOptions } from 'apexcharts';
import { LoadingPlaceHolder } from '../loading';
import { Box } from '../common-console-components';

// @ts-ignore
const ChartsMain = React.lazy(() => import('./charts'));

const Chart = (
  props: {
    options: ApexOptions;
  } & {
    title?: string;
    height?: string;
    width?: string;
    className?: string;
  }
) => {
  const { height, width, className, title } = props;
  return (
    <Suspense
      fallback={
        <div
          style={{
            width: width || '100%',
            height: height || '100%',
          }}
          className={cn(className, 'flex flex-col')}
        >
          <Box title={title} className="h-full">
            <div className="flex flex-col justify-center h-full">
              <LoadingPlaceHolder />
            </div>
          </Box>
        </div>
      }
    >
      <div
        className={cn(className, 'flex flex-col')}
        style={{
          height: height || '100%',
          width: width || '100%',
        }}
      >
        <Box title={title}>
          <ChartsMain {...props} />
        </Box>
      </div>
    </Suspense>
  );
};

export default Chart;
