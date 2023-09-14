import { List, SquaresFour } from '@jengaicons/react';
import {
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import Toolbar from '~/components/atoms/toolbar';
import { isBrowser } from '~/root/lib/client/helpers/is-browser';
import { IListOrGrid } from '../server/r-utils/common';

const ViewModeContext = createContext<{
  viewMode: IListOrGrid;
  setViewMode: (mode: IListOrGrid) => void;
}>({ viewMode: 'list', setViewMode() {} });

const retriveInitialViewMode = (): IListOrGrid => {
  return (
    isBrowser() ? sessionStorage.getItem('ViewMode') || 'list' : 'list'
  ) as IListOrGrid;
};

const saveViewMode = ({ mode }: { mode: IListOrGrid }) => {
  if (isBrowser()) {
    sessionStorage.setItem('ViewMode', mode);
  }
};

export const ViewModeProvider = ({ children }: { children: ReactNode }) => {
  const [viewMode, setViewMode] = useState(retriveInitialViewMode());

  useEffect(() => {
    saveViewMode({ mode: viewMode });
  }, [viewMode]);

  return (
    <ViewModeContext.Provider
      value={useMemo(
        () => ({ viewMode, setViewMode }),
        [viewMode, setViewMode]
      )}
    >
      {children}
    </ViewModeContext.Provider>
  );
};

export const useViewMode = () => {
  return useContext(ViewModeContext);
};

// Button for toggling between grid and list view
const ViewMode = () => {
  const { viewMode, setViewMode } = useViewMode();

  const [tempViewMode, setTempViewMode] = useState<IListOrGrid>('list');

  useEffect(() => {
    setTempViewMode(viewMode);
  }, [viewMode]);

  return (
    <Toolbar.ButtonGroup.Root
      value={tempViewMode}
      onValueChange={setViewMode}
      selectable
    >
      <Toolbar.ButtonGroup.IconButton icon={<List />} value="list" />
      <Toolbar.ButtonGroup.IconButton icon={<SquaresFour />} value="grid" />
    </Toolbar.ButtonGroup.Root>
  );
};

export default ViewMode;
