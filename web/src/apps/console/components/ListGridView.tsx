import { ReactNode, useEffect, useState } from 'react';
import { IListOrGrid } from '../server/r-utils/common';
import { useViewMode } from './view-mode';

const ListGridView = ({
  gridView,
  listView,
}: {
  gridView: ReactNode;
  listView: ReactNode;
}) => {
  const { viewMode } = useViewMode();
  const [tempViewMode, setTempViewMode] = useState<IListOrGrid>('list');

  useEffect(() => {
    setTempViewMode(viewMode);
  }, [viewMode]);

  return (
    <div>
      <div className="hidden md:block">
        {tempViewMode === 'grid' ? gridView : listView}
      </div>
      <div className="block md:hidden">{gridView}</div>
    </div>
  );
};

export default ListGridView;
