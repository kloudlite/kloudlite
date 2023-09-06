import { createContext, useContext, useEffect, useMemo } from 'react';
import { ImmerHook, useImmer } from 'use-immer';
import { ChildrenProps } from '~/components/types';
import { NonNullableString } from '~/root/lib/types/common';
import { AppIn } from '~/root/src/generated/gql/server';

const defaultApp: AppIn = {
  metadata: {
    name: '',
  },
  spec: {
    containers: [],
  },
  displayName: '',
};

type ISetState<T = any> = (fn: ((val: T) => T) | T) => void;

const CreateAppContext = createContext<any>(null);

export type createAppTabs =
  | 'environment'
  | 'application_details'
  | 'compute'
  | 'network'
  | 'review'
  | NonNullableString;

interface IappState {
  page: createAppTabs;
  app: AppIn;
}

export const useAppState = () => {
  const [state, setState] = useContext<ImmerHook<IappState>>(CreateAppContext);

  const { app, page } = state;

  const setApp: ISetState<typeof app> = (fn) => {
    if (typeof fn === 'function') {
      setState((s) => ({ ...s, app: fn(s.app) }));
    } else {
      setState((s) => ({ ...s, app: fn }));
    }
  };

  const setPage: ISetState<createAppTabs> = (fn) => {
    if (typeof fn === 'function') {
      setState((s) => ({ ...s, page: fn(s.page) }));
    } else {
      setState((s) => ({ ...s, page: fn }));
    }
  };

  useEffect(() => {
    if (!page) {
      setPage('application_details');
    }
    if (!app) {
      setApp(defaultApp);
    }
  }, []);

  return {
    app: app || defaultApp,
    setApp,
    page,
    setPage,
    state,
    setState,
  };
};

export const AppContextProvider = ({ children }: ChildrenProps) => {
  const loadSession = () => {
    if (typeof window === 'undefined')
      return {
        app: defaultApp,
      };
    const stateString =
      sessionStorage.getItem('state') ||
      JSON.stringify({
        app: defaultApp,
      });

    try {
      const data = JSON.parse(stateString);
      return data || {};
    } catch (_) {
      return {};
    }
  };
  const [state, setState] = useImmer<IappState>(() => {
    return loadSession();
  });

  useEffect(() => {
    if (typeof window === 'undefined') return;
    console.log(state);
    sessionStorage.setItem('state', JSON.stringify(state || {}));
  }, [state]);

  return (
    <CreateAppContext.Provider
      value={useMemo(() => [state, setState], [state, setState])}
    >
      {children}
    </CreateAppContext.Provider>
  );
};
