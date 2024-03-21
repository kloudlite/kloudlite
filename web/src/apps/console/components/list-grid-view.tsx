import { ReactNode } from 'react';
import { useViewMode } from './view-mode';

const ListGridView = ({
  gridView,
  listView,
}: {
  gridView: ReactNode;
  listView: ReactNode;
}) => {
  const { viewMode } = useViewMode();
  return (
    <div>
      <div className="hidden md:block">
        {viewMode === 'c' ? gridView : listView}
      </div>
      <div className="block md:hidden">{gridView}</div>
    </div>
  );
};

export default ListGridView;
