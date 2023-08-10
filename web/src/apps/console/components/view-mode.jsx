import { List, SquaresFour } from '@jengaicons/react';
import { useEffect, useState } from 'react';
import Toolbar from '~/components/atoms/toolbar';

// Button for toggling between grid and list view
const ViewMode = ({ mode, onModeChange }) => {
  const [m, setM] = useState(mode);
  useEffect(() => {
    if (onModeChange) onModeChange(m);
  }, [m]);
  return (
    <Toolbar.ButtonGroup value={m} onValueChange={setM} selectable>
      <Toolbar.ButtonGroup.IconButton icon={List} value="list" />
      <Toolbar.ButtonGroup.IconButton icon={SquaresFour} value="grid" />
    </Toolbar.ButtonGroup>
  );
};

export default ViewMode;
