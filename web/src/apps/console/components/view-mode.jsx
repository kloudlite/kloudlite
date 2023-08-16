import { List, SquaresFour } from '@jengaicons/react';
import { useEffect, useState } from 'react';
import Toolbar from '~/components/atoms/toolbar';

// Button for toggling between grid and list view
const ViewMode = ({ mode, onModeChange = (_) => _ }) => {
  const [m, setM] = useState(mode);
  useEffect(() => {
    onModeChange(m);
  }, [m]);
  return (
    <Toolbar.ButtonGroup.Root value={m} onValueChange={setM}>
      <Toolbar.ButtonGroup.IconButton icon={List} value="list" />
      <Toolbar.ButtonGroup.IconButton icon={SquaresFour} value="grid" />
    </Toolbar.ButtonGroup.Root>
  );
};

export default ViewMode;
