import { Smiley } from '@jengaicons/react';
import { EmptyState } from '~/console/components/empty-state';

const Wip = () => {
  return (
    <div className="py-4xl">
      <EmptyState heading="Comming Soon" image={<Smiley size={48} />} />
    </div>
  );
};

export default Wip;
