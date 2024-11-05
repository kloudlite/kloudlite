import { Smiley } from '~/console/components/icons';
import { EmptyState } from '~/console/components/empty-state';

const Wip = () => {
  return (
    <div className="py-4xl">
      <EmptyState heading="Coming Soon" image={<Smiley size={48} />} />
    </div>
  );
};

export default Wip;
