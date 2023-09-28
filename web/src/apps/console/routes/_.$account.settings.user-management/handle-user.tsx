import { useOutletContext } from '@remix-run/react';
import { TextInput } from '~/components/atoms/input';
import SelectPrimitive from '~/components/atoms/select-primitive';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ACCOUNT_ROLES } from '~/console/utils/commons';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IAccountContext } from '../_.$account';

const HandleUser = ({ show, setShow }: IDialog) => {
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
              userEmail: val.email,
              userRole: val.role as any,
            },
          });
          if (e) {
            throw e[0];
          }
          toast.success('user invited');
          setShow(null);
        } catch (err) {
          handleError(err);
        }
      },
    });

  return (
    <Popup.Root
      show={show as any}
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

            <SelectPrimitive.Root
              label="Role"
              value={values.role}
              onChange={handleChange('role')}
            >
              {Object.entries(ACCOUNT_ROLES).map(([key, value]) => {
                return (
                  <SelectPrimitive.Option key={value} value={key}>
                    {value}
                  </SelectPrimitive.Option>
                );
              })}
            </SelectPrimitive.Root>
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

export default HandleUser;
