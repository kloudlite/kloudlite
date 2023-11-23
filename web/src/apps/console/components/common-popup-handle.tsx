import { ReactNode } from 'react';
import Popup from '~/components/molecule/popup';
import { IDialogBase } from './types.d';

const CommonPopupHandle = <T,>({
  root,
  ...props
}: {
  root: (props: IDialogBase<T>) => ReactNode;
  updateTitle: ReactNode;
  createTitle: ReactNode;
} & IDialogBase<T>) => {
  const { isUpdate, visible, setVisible, createTitle, updateTitle } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>{isUpdate ? updateTitle : createTitle}</Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && root({ ...props })}
    </Popup.Root>
  );
};

export default CommonPopupHandle;
