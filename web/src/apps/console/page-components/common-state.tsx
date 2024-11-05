import { createContext, useContext, useEffect, useMemo } from 'react';
import { ImmerHook, useImmer } from 'use-immer';
import { ChildrenProps } from '@kloudlite/design-system/types';

const defaultData = {};
const stateName = 'common_config_data';

// type ISetState<T = any> = (fn: ((val: T) => T) | T) => void;

const CreateDataContext = createContext<any>(null);

interface IdataState {
  [key: string]: any;
}

export const useDataState = <T,>(key: string) => {
  const [state, setState] =
    useContext<ImmerHook<IdataState>>(CreateDataContext);
  const resetState = () => {
    setState(() => defaultData);
  };

  return {
    state: (state ? state[key] || {} : {}) as T,
    setState: (fn: ((val: T) => any) | T) => {
      if (typeof fn === 'function') {
        setState((draft: IdataState) => {
          // @ts-ignore
          draft[key] = fn(draft[key] || {});
        });
      } else {
        setState((draft: IdataState) => {
          draft[key] = fn;
        });
      }
    },
    resetState,
  };
};

export const clearDataState = () => {
  if (typeof window === 'undefined') return;
  sessionStorage.removeItem(stateName);
};

export const DataContextProvider = ({ children }: ChildrenProps) => {
  const loadSession = () => {
    if (typeof window === 'undefined') return defaultData;
    const stateString =
      sessionStorage.getItem(stateName) || JSON.stringify(defaultData);

    try {
      const data = JSON.parse(stateString);
      return data || {};
    } catch (_) {
      return {};
    }
  };
  const [state, setState] = useImmer<IdataState>(() => {
    return loadSession();
  });

  useEffect(() => {
    if (typeof window === 'undefined') return;
    sessionStorage.setItem(stateName, JSON.stringify(state || {}));
  }, [state]);

  return (
    <CreateDataContext.Provider
      value={useMemo(() => [state, setState], [state, setState])}
    >
      {children}
    </CreateDataContext.Provider>
  );
};
