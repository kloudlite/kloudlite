import { ProgressTracker } from '~/components/organisms/progress-tracker';
import { cn } from '~/components/utils';
import { Button } from '~/components/atoms/button';
import DevOpsImage from '../assets/dev-ops.png';

const Published = () => {
  return (
    <div
      className={cn(
        'transition-all flex justify-between',
        'flex-col md:flex-row',
        'py-6xl md:py-8xl',
        'gap-3xl md:gap-6xl lg:gap-9xl xl:gap-10xl'
      )}
    >
      <div className="flex flex-col gap-3xl items-start">
        <div className="flex flex-col gap-lg">
          <div className="heading4xl text-text-default">
            Congratulations! 🚀
          </div>
          <div className="text-text-soft">
            You just published a new project to Kloudlite.
          </div>
        </div>
        <ProgressTracker
          items={[
            {
              label: 'Configure project',
              active: true,
              key: 'configureproject',
            },
            {
              label: 'Review',
              active: true,
              key: 'review',
            },
          ]}
        />
      </div>
      <div className="flex flex-col gap-4xl">
        <div className="p-3xl flex-1 rounded-lg border border-border-default shadow-popover">
          <img src={DevOpsImage} alt="dev-ops" className="rounded" />
        </div>
        <Button variant="basic" content="Continue to dashboard" block />
      </div>
    </div>
  );
};

export default Published;
