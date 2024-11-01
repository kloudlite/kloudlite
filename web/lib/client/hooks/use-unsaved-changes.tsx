/* eslint-disable camelcase */
import {
  unstable_useBlocker,
  useBeforeUnload,
  useLocation,
  useRevalidator,
} from '@remix-run/react';
import {
  ReactNode,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import Popup from '@kloudlite/design-system/molecule/popup';
import { useReload } from '../helpers/reloader';

const UnsavedChanges = createContext<{
  hasChanges: boolean;
  setHasChanges: (hasChanges: boolean) => void;
  unloadState: 'unblocked' | 'blocked' | 'proceeding';
  proceed: (() => void) | undefined;
  reset: (() => void) | undefined;
  resetAndReload: () => void;
  setIgnorePaths: (path: string[]) => void;
  performAction: string;
  setPerformAction: (action: string) => void;
  loading: boolean;
}>({
  hasChanges: false,
  setHasChanges() {},
  unloadState: 'unblocked',
  proceed() {},
  reset() {},
  resetAndReload() {},
  setIgnorePaths() {},
  performAction: '',
  setPerformAction() {},
  loading: false,
});

export const UnsavedChangesProvider = ({
  children,
  onProceed,
}: {
  children?: ReactNode;
  onProceed?: (props: { setPerformAction?: (action: string) => void }) => void;
}) => {
  const [hasChanges, setHasChanges] = useState<boolean>(false);
  const [reload, setReload] = useState(false);
  const [ignorePaths, setIgnorePaths] = useState<string[]>([]);
  const [performAction, setPerformAction] = useState<string>('');
  const location = useLocation();
  const { state, proceed, reset } = unstable_useBlocker(({ nextLocation }) => {
    if (hasChanges && !ignorePaths.includes(nextLocation.pathname)) {
      return true;
    }
    return false;
  });

  useBeforeUnload(
    useCallback(
      (e) => {
        if (hasChanges) {
          e.preventDefault();
        }
        return hasChanges;
      },
      [hasChanges]
    )
  );

  useEffect(() => {
    if (
      !(
        ignorePaths &&
        ignorePaths.length > 0 &&
        ignorePaths.includes(location.pathname)
      )
    ) {
      setHasChanges(false);
    }
  }, [location]);

  const refresh = useReload();

  useEffect(() => {
    if (reload && !hasChanges) {
      refresh();
      setReload(false);
    }
  }, [reload, hasChanges]);

  const resetAndReload = () => {
    setHasChanges(false);
    setReload(true);
  };

  const { state: s } = useRevalidator();
  return (
    <UnsavedChanges.Provider
      value={useMemo(
        () => ({
          hasChanges,
          setHasChanges,
          unloadState: state,
          proceed,
          reset,
          resetAndReload,
          setIgnorePaths,
          setPerformAction,
          performAction,
          loading: s === 'loading',
        }),
        [
          hasChanges,
          setHasChanges,
          state,
          ignorePaths,
          performAction,
          setPerformAction,
          s,
        ]
      )}
    >
      {children}
      <Popup.Root
        show={state === 'blocked'}
        onOpenChange={() => {
          reset?.();
        }}
      >
        <Popup.Header>Unsaved changes</Popup.Header>
        <Popup.Content>
          Are you sure you want to discard the changes?
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button
            content="Cancel"
            variant="basic"
            onClick={() => reset?.()}
          />
          <Popup.Button
            content="Discard"
            variant="warning"
            onClick={() => {
              proceed?.();
              onProceed?.({ setPerformAction });
            }}
          />
        </Popup.Footer>
      </Popup.Root>
    </UnsavedChanges.Provider>
  );
};

export const DISCARD_ACTIONS = {
  DISCARD_CHANGES: 'discard-changes',
  VIEW_CHANGES: 'view-changes',
  INIT: 'init',
};

export const useUnsavedChanges = () => {
  return useContext(UnsavedChanges);
};
