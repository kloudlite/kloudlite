import React, { Suspense } from 'react';
import { cn } from '~/components/utils';
import { ChartProps } from './charts';
import { LoadingPlaceHolder } from '../loading';

// @ts-ignore
const ChartsMain = React.lazy(() => import('./charts'));

const Chart = (
  props: ChartProps & {
    height?: string;
    width?: string;
    className?: string;
  }
) => {
  const { height, width, className } = props;
  return (
    <Suspense
      fallback={
        <div
          style={{
            width: width || '100%',
            height: height || '100%',
          }}
          className={cn(
            'flex justify-center items-center bg-text-primary-100',
            className
          )}
        >
          <LoadingPlaceHolder />
        </div>
      }
    >
      <div
        className={cn(className)}
        style={{
          height: height || '100%',
          width: width || '100%',
        }}
      >
        <ChartsMain {...props} />
      </div>
    </Suspense>
  );
};

export default Chart;
