/* eslint-disable camelcase */
import {
  unstable_useBlocker,
  useBeforeUnload,
  useLocation,
} from '@remix-run/react';
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { ChildrenProps } from '~/components/types';
import Popup from '~/components/molecule/popup';
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
  setLoading: (loading: boolean) => void;
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
  setLoading() {},
});

export const UnsavedChangesProvider = ({ children }: ChildrenProps) => {
  const [hasChanges, setHasChanges] = useState<boolean>(false);
  const [reload, setReload] = useState(false);
  const [ignorePaths, setIgnorePaths] = useState<string[]>([]);
  const [performAction, setPerformAction] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
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
          loading,
          setLoading,
        }),
        [
          hasChanges,
          setHasChanges,
          state,
          ignorePaths,
          performAction,
          setPerformAction,
          loading,
          setLoading,
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
            onClick={() => proceed?.()}
          />
        </Popup.Footer>
      </Popup.Root>
    </UnsavedChanges.Provider>
  );
};
export const useUnsavedChanges = () => {
  return useContext(UnsavedChanges);
};
