import SelectInput from '~/components/atoms/select';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useOutletContext } from '@remix-run/react';
import { toast } from '~/components/molecule/toast';
import { IHandleProps } from '~/console/server/utils/common';
import { IAccountContext } from '../_.$account';

const roles = Object.freeze({
  member: 'account-member',
  admin: 'account-admin',
});

const Main = ({ show, setShow }: IHandleProps) => {
  const api = useConsoleApi();

  const { account } = useOutletContext<IAccountContext>();

  const { values, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        email: '',
        role: 'account-member',
      },
      validationSchema: Yup.object({
        email: Yup.string().required().email(),
      }),
      onSubmit: async (val) => {
        try {
          const { errors: e } = await api.inviteMemberForAccount({
            accountName: account.metadata.name,
            invitation: {
              accountName: account.metadata.name,
              userEmail: val.email,
              userRole: val.role,
            },
          });
          if (e) {
            throw e[0];
          }
          toast.success('user invited');
          setShow(false);
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
              onChange={handleChange('role')}
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
          <Popup.Button
            loading={isLoading}
            type="submit"
            content="Send invite"
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

const HandleUser = ({ show, setShow }: IHandleProps) => {
  if (show) {
    return <Main show={show} setShow={setShow} />;
  }
  return null;
};

export default HandleUser;
