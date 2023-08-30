import { useParams } from '@remix-run/react';
import SelectInput from '~/components/atoms/select';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/types/common';

const roles = Object.freeze({
  member: 'account-member',
  admin: 'account-admin',
});

const Main = ({ show, setShow }) => {
  const api = useAPIClient();

  const { account } = useParams();

  const { values, handleChange, handleSubmit, resetValues } = useForm({
    initialValues: {
      email: '',
    },
    validationSchema: Yup.object({
      email: Yup.string().required().email(),
    }),
    onSubmit: async (val) => {
      try {
        const { errors: e } = await api.inviteUser({
          accountName: account,
          email: val.email,
          role: val.role,
        });
        if (e) {
          throw e[0];
        }
      } catch (err) {
        handleError(err);
      }
    },
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
          <div className="flex gap-2xl">
            <div className="flex-1">
              <TextInput
                label="Email"
                value={values.email}
                onChange={handleChange('email')}
              />
            </div>

            <SelectInput.Root
              label="Role"
              value={values.role}
              size="lg"
              onChange={(v) => {
                handleChange('role')({ target: { value: v } });
              }}
            >
              <SelectInput.Option> -- not-selected -- </SelectInput.Option>
              {[roles.admin, roles.member].map((role) => {
                return (
                  <SelectInput.Option key={role} value={role}>
                    {role}
                  </SelectInput.Option>
                );
              })}
            </SelectInput.Root>
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
