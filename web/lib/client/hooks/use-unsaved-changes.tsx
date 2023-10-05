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
import { useReload } from '../helpers/reloader';

const UnsavedChanges = createContext<{
  hasChanges: boolean;
  setHasChanges: (hasChanges: boolean) => void;
  unloadState: 'unblocked' | 'blocked' | 'proceeding';
  proceed: (() => void) | undefined;
  reset: (() => void) | undefined;
  resetAndReload: () => void;
}>({
  hasChanges: false,
  setHasChanges() {},
  unloadState: 'unblocked',
  proceed() {},
  reset() {},
  resetAndReload() {},
});

export const UnsavedChangesProvider = ({ children }: ChildrenProps) => {
  const [hasChanges, setHasChanges] = useState<boolean>(false);
  const [reload, setReload] = useState(false);
  const { state, proceed, reset } = unstable_useBlocker(hasChanges);
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
  const location = useLocation();

  useEffect(() => {
    setHasChanges(false);
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
        }),
        [hasChanges, setHasChanges, state]
      )}
    >
      {children}
    </UnsavedChanges.Provider>
  );
};
export const useUnsavedChanges = () => {
  return useContext(UnsavedChanges);
};
