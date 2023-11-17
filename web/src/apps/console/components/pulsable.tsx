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
          light: '#bebebe82',
          medium: '#bebebe82',
        },
      }}
      isLoading={isLoading}
    >
      {children}
    </ReactPulsable>
  );
};

export default Pulsable;
