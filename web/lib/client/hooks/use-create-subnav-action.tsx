import { useLocation } from '@remix-run/react';
import {
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { ChildrenProps } from '~/components/types';

const SubNavDataContext = createContext<{
  data?: ReactNode;
  setData: (data: ReactNode) => void;
}>({
  setData() {},
});

export const SubNavDataProvider = ({ children }: ChildrenProps) => {
  const [data, setData] = useState<ReactNode | undefined>();

  const location = useLocation();
  useEffect(() => {
    setData(undefined);
  }, [location.pathname]);

  return (
    <SubNavDataContext.Provider
      value={useMemo(
        () => ({
          data,
          setData,
        }),
        [data, setData]
      )}
    >
      {children}
    </SubNavDataContext.Provider>
  );
};
export const useSubNavData = () => {
  return useContext(SubNavDataContext);
};
