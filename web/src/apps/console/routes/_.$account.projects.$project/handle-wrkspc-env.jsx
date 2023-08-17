import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import * as Popup from '~/components/molecule/popup';
import {
  BlackProdLogo,
  BlackWorkspaceLogo,
} from '~/console/components/commons';
import { IdSelector } from '~/console/components/id-selector';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';

const Main = ({ show, setShow }) => {
  const { values, handleChange, handleSubmit, resetValues } = useForm({
    initialValues: {
      name: '',
    },
    validationSchema: Yup.object({}),
    onSubmit: async () => {},
  });

  return (
    <Popup.PopupRoot
      show={show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header showclose={false}>
        <div className="flex flex-row gap-lg items-center">
          {' '}
          {show?.type === 'workspaces' ? (
            <>
              <BlackWorkspaceLogo size={28} /> New Workspace
            </>
          ) : (
            <>
              <BlackProdLogo size={28} />
              New Environment
            </>
          )}
        </div>
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Name"
              value={values.name}
              onChange={handleChange('name')}
            />
            <IdSelector name={values.name} />
            <Select.Root label="Domains" />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button type="submit" content="Create" variant="primary" />
        </Popup.Footer>
      </form>
    </Popup.PopupRoot>
  );
};

export const HandlePopup = ({ show, setShow }) => {
  if (show) {
    return <Main show={show} setShow={setShow} />;
  }
  return null;
};
