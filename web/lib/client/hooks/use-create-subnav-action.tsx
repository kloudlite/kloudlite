import { createContext, useContext, useMemo, useState } from 'react';
import { ChildrenProps } from '~/components/types';
import { ISubNavCallback } from '~/console/components/types.d';

const SubNavDataContext = createContext<{
  data?: ISubNavCallback;
  setData: (data: ISubNavCallback) => void;
}>({
  setData() {},
});

export const SubNavDataProvider = ({ children }: ChildrenProps) => {
  const [data, setData] = useState<ISubNavCallback | undefined>();

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
