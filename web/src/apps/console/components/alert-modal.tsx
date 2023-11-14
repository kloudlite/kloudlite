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
  okDisabled = false,
  okText = 'Delete',
  cancelText = 'Cancel',
  variant = 'critical',
}: IAlertModal) => {
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
      <form onSubmit={handleSubmit}>
        <AlertDialog.Content>{message}</AlertDialog.Content>
        {footer && (
          <AlertDialog.Footer>
            <AlertDialog.Button variant="basic" content={cancelText} closable />
            <AlertDialog.Button
              type="submit"
              disabled={okDisabled}
              variant={variant}
              content={okText}
              closable={false}
              loading={isLoading}
            />
          </AlertDialog.Footer>
        )}
      </form>
    </AlertDialog.Root>
  );
};

export default AlertModal;
