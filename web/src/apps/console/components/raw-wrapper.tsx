import { Key, ReactNode } from 'react';
import { Button } from '~/components/atoms/button';
import Tooltip from '~/components/atoms/tooltip';
import { BrandLogo } from '~/components/branding/brand-logo';
import ProgressTracker, {
  ProgressItemProps,
} from '~/components/organisms/progress-tracker';
import { cn } from '~/components/utils';

interface IRawWrapper<I = any, V = any, C = number | string> {
  title: string;
  subtitle: string;
  badge?: {
    title?: string;
    subtitle?: string;
    image?: ReactNode;
  };
  progressItems?: ProgressItemProps<I & { id: C; label: ReactNode }, V>[];
  onProgressClick?: (value: V) => void;
  onCancel?: () => void;
  rightChildren: ReactNode;
}
function RawWrapper<I = any, V = any, C = number | string>({
  title,
  subtitle,
  progressItems,
  onProgressClick = () => {},
  onCancel,
  badge,
  rightChildren,
}: IRawWrapper<I, V, C>) {
  return (
    <Tooltip.Provider>
      <div className="min-h-screen flex flex-row">
        <div className="min-h-full flex flex-col bg-surface-basic-subdued px-11xl pt-11xl pb-10xl">
          <div className="flex flex-col items-start gap-6xl w-[379px]">
            <BrandLogo detailed={false} size={48} />
            <div
              className={cn('flex flex-col', {
                'gap-8xl': !!badge?.title || !!badge?.subtitle,
                'gap-4xl': !badge?.title && !badge?.subtitle,
              })}
            >
              <div className="flex flex-col gap-3xl">
                <div className="text-text-default heading4xl">{title}</div>
                <div className="text-text-default bodyLg">{subtitle}</div>
                {(!!badge?.title || !!badge?.subtitle) && (
                  <div className="flex flex-row gap-lg p-lg rounded border border-border-default bg-surface-basic-active min-w-[120px] w-fit">
                    {badge.image && (
                      <div className="p-md text-icon-default flex items-center rounded bg-surface-basic-default">
                        {badge?.image}
                      </div>
                    )}
                    <div className="flex flex-col">
                      <div className="bodySm-semibold text-text-default">
                        {badge?.title}
                      </div>
                      <div className="bodySm text-text-soft">
                        {badge?.subtitle}
                      </div>
                    </div>
                  </div>
                )}
              </div>
              {progressItems && (
                <ProgressTracker.Root
                  items={progressItems}
                  onClick={(v) => {
                    onProgressClick(v);
                  }}
                >
                  {(item) => {
                    return (
                      <ProgressTracker.Item
                        key={item.id as Key}
                        active={item.active}
                        completed={item.completed}
                      >
                        {item.label}
                      </ProgressTracker.Item>
                    );
                  }}
                </ProgressTracker.Root>
              )}
            </div>

            {!!onCancel && (
              <Button
                variant="outline"
                content="Cancel"
                size="lg"
                onClick={onCancel}
              />
            )}
          </div>
        </div>
        <div className="pt-11xl pb-12xl px-11xl flex flex-1 bg-surface-basic-default">
          <div className="w-[628px] flex items-center">
            <div className="flex flex-col gap-6xl w-full">{rightChildren}</div>
          </div>
        </div>
      </div>
    </Tooltip.Provider>
  );
}

interface ITitleBox {
  title: ReactNode;
  subtitle: ReactNode;
}
export const TitleBox = ({ title, subtitle }: ITitleBox) => {
  return (
    <div className="flex flex-col gap-lg">
      <div className="headingXl text-text-default">{title}</div>
      {subtitle && <div className="bodyMd text-text-soft">{subtitle}</div>}
    </div>
  );
};

export default RawWrapper;
