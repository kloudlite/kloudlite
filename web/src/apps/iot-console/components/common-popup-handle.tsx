import { ComponentType, ReactNode } from 'react';
import Popup from '~/components/molecule/popup';
import { IDialogBase } from './types.d';

const CommonPopupHandle = <T,>({
  root: Root,
  ...props
}: {
  root: ComponentType<IDialogBase<T>>;
  updateTitle: ReactNode;
  createTitle: ReactNode;
} & IDialogBase<T>) => {
  const { isUpdate, visible, setVisible, createTitle, updateTitle } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>{isUpdate ? updateTitle : createTitle}</Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default CommonPopupHandle;
