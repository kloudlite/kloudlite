import ReactPulsable from 'react-pulsable';
import { ChildrenProps } from '~/components/types';

const Pulsable = ({
  children,
  isLoading,
}: ChildrenProps & { isLoading: boolean }) => {
  return (
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
  );
};

export default Pulsable;
