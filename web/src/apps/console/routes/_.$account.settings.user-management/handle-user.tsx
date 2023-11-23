import { useOutletContext } from '@remix-run/react';
import { TextInput } from '~/components/atoms/input';
import SelectPrimitive from '~/components/atoms/select-primitive';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { Kloudlite__Io___Apps___Iam___Types__Role as Role } from '~/root/src/generated/gql/server';
import { parseName } from '~/console/server/r-utils/common';
import { IAccountContext } from '../_.$account';

const validRoles = (role: string): Role => {
  switch (role as Role) {
    case 'account_owner':
    case 'account_admin':
    case 'account_member':
      return role as Role;
    default:
      throw new Error(`invalid role ${role}`);
  }
};

const HandleUser = ({ show, setShow }: IDialog) => {
  const api = useConsoleApi();

  const { account } = useOutletContext<IAccountContext>();

  const { values, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        email: '',
        role: 'account_member',
      },
      validationSchema: Yup.object({
        email: Yup.string().required().email(),
      }),
      onSubmit: async (val) => {
        try {
          const { errors: e } = await api.inviteMembersForAccount({
            accountName: parseName(account),
            invitations: {
              userEmail: val.email,
              userRole: validRoles(val.role),
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

  const roles: Role[] = ['account_owner', 'account_admin', 'account_member'];

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
              <SelectPrimitive.Option value="">
                -- not-selected --
              </SelectPrimitive.Option>
              {roles.map((role) => {
                return (
                  <SelectPrimitive.Option key={role} value={role}>
                    {role}
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
