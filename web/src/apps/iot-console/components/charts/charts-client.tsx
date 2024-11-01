import React, { Suspense } from 'react';
import { cn } from '@kloudlite/design-system/utils';
import { type ApexOptions } from 'apexcharts';
import { LoadingPlaceHolder } from '../loading';
import { Box } from '../common-console-components';

// @ts-ignore
const ChartsMain = React.lazy(() => import('./charts'));

const Chart = (
  props: {
    options: ApexOptions;
  } & {
    disabled?: boolean;
    title?: string;
    height?: string;
    width?: string;
    className?: string;
  }
) => {
  const { height, width, className, title, disabled } = props;
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
        <div
          className={cn(
            'rounded border border-border-default bg-surface-basic-default shadow-button flex flex-col',
            className
          )}
        >
          <div className="text-text-strong headingMd px-2xl pt-2xl">
            {title}
          </div>
          <div className="relative">
            {disabled && (
              <div className="flex absolute inset-0 justify-center items-center bg-surface-basic-default z-10 headingSm">
                Not Available
              </div>
            )}

            <ChartsMain {...props} />
          </div>
        </div>
      </div>
    </Suspense>
  );
};

export default Chart;
