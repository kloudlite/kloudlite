import { useOutletContext } from '@remix-run/react';
import { TextInput } from '@kloudlite/design-system/atoms/input';
import Popup from '@kloudlite/design-system/molecule/popup';
import { toast } from '@kloudlite/design-system/molecule/toast';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { IDialogBase } from '~/console/components/types.d';
import { IMemberType } from '~/console/routes/_main+/$account+/settings+/user-management/user-access-resource';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { Github__Com___Kloudlite___Api___Apps___Iam___Types__Role as Role } from '~/root/src/generated/gql/server';
import { IAccountContext } from '../../_layout';

type IDialog = IDialogBase<IMemberType>;

const validRoles = (role: string): Role => {
  switch (role) {
    case 'account_owner':
      return 'account_owner' as Role;
    case 'account_member':
      return 'account_member' as Role;
    default:
      throw new Error(`invalid role ${role}`);
  }
};

const Root = (props: IDialog) => {
  const { setVisible, isUpdate } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { account } = useOutletContext<IAccountContext>();

  const { values, handleChange, handleSubmit, isLoading } = useForm({
    initialValues: isUpdate
      ? {
          email: props?.data.email || '',
          // role: props?.data.role || 'account_member',
          role: 'account_member',
        }
      : {
          email: '',
          role: 'account_member',
        },
    validationSchema: Yup.object({
      email: Yup.string().required().email(),
    }),
    onSubmit: async (val) => {
      try {
        if (!isUpdate) {
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
        } else if (isUpdate) {
          const { errors: e } = await api.updateAccountMembership({
            accountName: parseName(account),
            memberId: props?.data.id,
            role: validRoles(val.role),
          });
          if (e) {
            throw e[0];
          }
        }
        reloadPage();
        toast.success(`user ${isUpdate ? 'role updated' : 'invited'}`);
        setVisible(false);
      } catch (err) {
        handleError(err);
      }
    },
  });

  const roles: string[] = ['account_owner', 'account_member'];

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <div className="flex gap-2xl">
          <div className="flex-1">
            {!isUpdate ? (
              <TextInput
                label="Email"
                value={values.email}
                onChange={handleChange('email')}
              />
            ) : (
              <TextInput label="Email" value={values.email} disabled />
            )}
          </div>

          {/* <SelectPrimitive.Root
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
                  {mapRoleToDisplayName(role)}
                </SelectPrimitive.Option>
              );
            })}
          </SelectPrimitive.Root> */}
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content={isUpdate ? 'update' : 'Send invite'}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleUser = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      createTitle="Invite Member"
      updateTitle="Update Member"
      root={Root}
    />
  );
};

export default HandleUser;
