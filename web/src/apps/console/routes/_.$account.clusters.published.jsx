import { cn } from '~/components/utils';
import { Button } from '~/components/atoms/button';
import ProgressTracker from '~/components/organisms/progress-tracker';
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
            You just published a new cluster to Kloudlite.
          </div>
        </div>
        <ProgressTracker.Root>
          {[
            {
              label: 'Configure cluster',
              active: true,
              id: 'configurecluster',
              completed: false,
            },
            {
              label: 'review',
              active: true,
              id: 'review',
              completed: false,
            },
          ].map((pi) => (
            <ProgressTracker.Item key={pi.id} {...pi} />
          ))}
        </ProgressTracker.Root>
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
