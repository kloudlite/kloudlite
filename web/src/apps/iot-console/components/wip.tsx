import { EmptyState } from '~/iotconsole/components/empty-state';
import { Smiley } from '~/iotconsole/components/icons';

const Wip = () => {
  return (
    <div className="py-4xl">
      <EmptyState heading="Coming Soon" image={<Smiley size={48} />} />
    </div>
  );
};

export default Wip;
