import ReactPulsable from 'react-pulsable';
import { ChildrenProps } from '~/components/types';

const Pulsable = ({
  children,
  isLoading,
}: ChildrenProps & { isLoading: boolean }) => {
  return (
    <ReactPulsable backgroundColor="#bebebe82" isLoading={isLoading}>
      {children}
    </ReactPulsable>
  );
};

export default Pulsable;
