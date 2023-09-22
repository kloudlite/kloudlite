import { ReactNode, useEffect } from 'react';
import { ButtonVariants } from '~/components/atoms/button';
import AlertDialog from '~/components/molecule/alert-dialog';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';

export interface IAlertModal {
  show: any;
  setShow: (show: any) => void;
  onSubmit?: (val: any) => any;
  setLoading?: (loading: boolean) => void;
  submitType?: 'button' | 'submit';
  message: ReactNode;
  title: ReactNode;
  okText?: string;
  okDisabled?: boolean;
  cancelText?: string;
  variant?: ButtonVariants;
  footer?: boolean;
}

const AlertModal = ({
  show = false,
  setShow = () => {},
  footer = true,
  onSubmit,
  setLoading,
  message,
  title,
  submitType = 'button',
  okDisabled = false,
  okText = 'Delete',
  cancelText = 'Cancel',
  variant = 'critical',
}: IAlertModal) => {
  let FormElement: any = 'form';
  if (submitType === 'button') {
    FormElement = 'div';
  }

  const { handleSubmit, isLoading } = useForm({
    initialValues: {},
    validationSchema: Yup.object({}),
    onSubmit: async () => {
      if (onSubmit) {
        await onSubmit(show);
      }
    },
  });

  useEffect(() => {
    if (setLoading) {
      setLoading(isLoading);
    }
  }, [isLoading]);

  return (
    <AlertDialog.Root show={show} onOpenChange={setShow}>
      <AlertDialog.Header>{title}</AlertDialog.Header>
      <FormElement
        {...(submitType === 'submit' ? { onSubmit: handleSubmit } : {})}
      >
        <AlertDialog.Content>{message}</AlertDialog.Content>
        {footer && (
          <AlertDialog.Footer>
            <AlertDialog.Button variant="basic" content={cancelText} closable />
            <AlertDialog.Button
              type={submitType}
              disabled={okDisabled}
              variant={variant}
              content={okText}
              closable={false}
              loading={isLoading}
            />
          </AlertDialog.Footer>
        )}
      </FormElement>
    </AlertDialog.Root>
  );
};

export default AlertModal;
