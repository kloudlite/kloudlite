import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';

const Main = ({ show, setShow }) => {
  const { values, handleChange, handleSubmit, resetValues } = useForm({
    initialValues: {
      email: '',
    },
    validationSchema: Yup.object({}),
    onSubmit: async () => {},
  });

  return (
    <Popup.Root
      show={show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>Invite user</Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Email"
              value={values.email}
              onChange={handleChange('email')}
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button type="submit" content="Send invite" variant="primary" />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

const HandleUser = ({ show, setShow }) => {
  if (show) {
    return <Main show={show} setShow={setShow} />;
  }
  return null;
};

export default HandleUser;
