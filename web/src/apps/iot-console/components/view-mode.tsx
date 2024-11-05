import {
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import Toolbar from '@kloudlite/design-system/atoms/toolbar';
import { isBrowser } from '~/root/lib/client/helpers/is-browser';
import { List, SquaresFour } from '~/iotconsole/components/icons';
import { IListOrGrid } from '../server/r-utils/common';

const ViewModeContext = createContext<{
  viewMode: IListOrGrid;
  setViewMode: (mode: IListOrGrid) => void;
}>({ viewMode: 'r', setViewMode() {} });

const setVM = ({ mode }: { mode: IListOrGrid }) => {
  sessionStorage.setItem('view-mode', mode);
};

const getVM = (): IListOrGrid => {
  if (isBrowser()) {
    try {
      return (sessionStorage.getItem('view-mode') || 'r') as IListOrGrid;
    } catch (err) {
      //
    }
  }
  return 'r';
};

export const ViewModeProvider = ({ children }: { children: ReactNode }) => {
  const [viewMode, setViewMode] = useState<IListOrGrid>(getVM());

  useEffect(() => {
    setVM({ mode: viewMode });
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

  return (
    <Toolbar.ButtonGroup.Root
      value={viewMode}
      onValueChange={setViewMode}
      selectable
    >
      <Toolbar.ButtonGroup.IconButton icon={<List />} value="r" />
      <Toolbar.ButtonGroup.IconButton icon={<SquaresFour />} value="c" />
    </Toolbar.ButtonGroup.Root>
  );
};

export default ViewMode;
