import { ReactNode } from 'react';
import { ButtonVariants } from '~/components/atoms/button';
import * as AD from '~/components/molecule/alert-dialog';

interface IAlertDialog {
  show: boolean;
  setShow: (show: boolean) => void;
  onSubmit?: (val: any) => any;
  message: string;
  title: ReactNode;
  okText: ReactNode;
  type: ButtonVariants;
}

const AlertDialog = ({
  show = false,
  setShow = () => {},
  onSubmit,
  message,
  title,
  okText,
  type,
}: IAlertDialog) => {
  return (
    <AD.DialogRoot show={show} onOpenChange={setShow}>
      <AD.Header>{title}</AD.Header>
      <AD.Content>{message}</AD.Content>
      <AD.Footer>
        <AD.Button variant="basic" content="Cancel" />
        <AD.Button
          variant={type}
          content={okText}
          onClick={(e) => {
            e.preventDefault();
            if (onSubmit) onSubmit(show);
          }}
        />
      </AD.Footer>
    </AD.DialogRoot>
  );
};

export default AlertDialog;
