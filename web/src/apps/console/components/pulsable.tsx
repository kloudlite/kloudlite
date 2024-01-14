import { createContext, useContext, useMemo } from 'react';
import ReactPulsable from 'react-pulsable';
import { ChildrenProps } from '~/components/types';

const pulsableContext = createContext(false);

export const usePulsableLoading = () => {
  return useContext(pulsableContext);
};

const Pulsable = ({
  children,
  isLoading,
}: ChildrenProps & { isLoading: boolean }) => {
  return (
    <pulsableContext.Provider value={useMemo(() => isLoading, [isLoading])}>
      <ReactPulsable
        config={{
          bgColors: {
            light: 'rgba(161, 161, 170, 0.2)',
            medium: 'rgba(161, 161, 170, 0.3)',
          },
        }}
        isLoading={isLoading}
      >
        {children}
      </ReactPulsable>
    </pulsableContext.Provider>
  );
};

export default Pulsable;
