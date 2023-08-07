import * as AD from '~/components/molecule/alert-dialog';

const AlertDialog = ({
  show,
  setShow,
  onSubmit,
  message,
  title,
  okText,
  type,
}) => {
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
