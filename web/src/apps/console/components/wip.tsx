import { Integration } from '@jengaicons/react';
import { EmptyState } from '~/console/components/empty-state';

const Wip = () => {
  return (
    <div className="py-4xl">
      <EmptyState
        heading="Page is under construction"
        image={<Integration size={48} />}
      />
    </div>
  );
};

export default Wip;
