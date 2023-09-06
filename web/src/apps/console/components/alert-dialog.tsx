import { ReactNode } from 'react';
import { ButtonVariants } from '~/components/atoms/button';
import * as AD from '~/components/molecule/alert-dialog';

export interface IAlertDialog {
  show: any;
  setShow: (show: any) => void;
  onSubmit?: (val: any) => any;
  message: ReactNode;
  title: ReactNode;
  okText?: string;
  cancelText?: string;
  type?: ButtonVariants;
  footer?: boolean;
}

const AlertDialog = ({
  show = false,
  setShow = () => {},
  footer = true,
  onSubmit,
  message,
  title,
  okText = 'Delete',
  cancelText = 'Cancel',
  type = 'critical',
}: IAlertDialog) => {
  return (
    <AD.DialogRoot show={show} onOpenChange={setShow}>
      <AD.Header>{title}</AD.Header>
      <AD.Content>{message}</AD.Content>
      {footer && (
        <AD.Footer>
          <AD.Button variant="basic" content={cancelText} />
          <AD.Button
            variant={type}
            content={okText}
            onClick={(e) => {
              e.preventDefault();
              if (onSubmit) onSubmit(show);
            }}
          />
        </AD.Footer>
      )}
    </AD.DialogRoot>
  );
};

export default AlertDialog;
