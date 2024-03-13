import { PencilSimple, Trash } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import { Avatar } from '~/components/atoms/avatar';
import { toast } from '~/components/molecule/toast';
import { titleCase } from '~/components/utils';
import {
  ListBody,
  ListTitle,
} from '~/console/components/console-list-components';
import DeleteDialog from '~/console/components/delete-dialog';
import List from '~/console/components/list';
import ListGridView from '~/console/components/list-grid-view';
import ResourceExtraAction, {
  IResourceExtraItem,
} from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { parseName } from '~/console/server/r-utils/common';
import HandleUser from '~/console/routes/_main+/$account+/settings+/user-management/handle-user';
import { IAccountContext } from '../../_layout';

const RESOURCE_NAME = 'user';

type BaseType = {
  id: string;
  name: string;
  role: string;
  email: string;
};
export type IMemberType = BaseType;

type OnAction = ({
  action,
  item,
}: {
  action: 'delete' | 'edit';
  item: BaseType;
}) => void;

type IExtraButton = {
  onAction: OnAction;
  item: BaseType;
  isInvite: boolean;
};

export const mapRoleToDisplayName = (role: string): string => {
  switch (role) {
    case 'account_owner':
      return 'owner';
    case 'account_member':
      return 'member';
    default:
      return role;
  }
};

const ExtraButton = ({ onAction, item, isInvite }: IExtraButton) => {
  let items: IResourceExtraItem[] = [
    {
      label: 'Remove',
      icon: <Trash size={16} />,
      type: 'item',
      onClick: () => onAction({ action: 'delete', item }),
      key: 'remove',
      className: '!text-text-critical',
    },
  ];
  if (!isInvite) {
    items = [
      {
        label: 'Edit',
        icon: <PencilSimple size={16} />,
        type: 'item',
        onClick: () => onAction({ action: 'edit', item }),
        key: 'edit',
      },
      ...items,
    ];
  }
  return <ResourceExtraAction options={items} />;
};

interface IResource {
  items: BaseType[];
  onAction: OnAction;
  isInvite: boolean;
}

const ListView = ({ items = [], onAction, isInvite }: IResource) => {
  return (
    <List.Root>
      {items.map((item) => (
        <List.Row
          key={item.id}
          className="!p-3xl"
          columns={[
            {
              key: 1,
              className: 'flex-1',
              render: () => (
                <ListTitle
                  avatar={<Avatar size="sm" />}
                  subtitle={item.email}
                  title={item.name}
                />
              ),
            },
            {
              key: 2,
              render: () => <ListBody data={mapRoleToDisplayName(item.role)} />,
            },
            {
              key: 3,
              render: () => (
                <ExtraButton
                  isInvite={isInvite}
                  onAction={onAction}
                  item={item}
                />
              ),
            },
          ]}
        />
      ))}
    </List.Root>
  );
};

const UserAccessResources = ({
  items = [],
  isPendingInvitation = false,
}: {
  items: BaseType[];
  isPendingInvitation: boolean;
}) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );
  const [showUserInvite, setShowUserInvite] = useState<BaseType | null>(null);

  const { account } = useOutletContext<IAccountContext>();

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    isInvite: isPendingInvitation,
    onAction: ({ action, item }) => {
      switch (action) {
        case 'edit':
          setShowUserInvite(item);
          break;
        case 'delete':
          setShowDeleteDialog(item);
          break;
        default:
          break;
      }
    },
  };

  return (
    <>
      <ListGridView
        listView={<ListView {...props} />}
        gridView={<ListView {...props} />}
      />
      <HandleUser
        {...{
          isUpdate: true,
          data: showUserInvite!,
          setVisible: () => setShowUserInvite(null),
          visible: !!showUserInvite,
        }}
      />
      <DeleteDialog
        resourceName="confirm"
        customMessages={{
          action: 'Remove',
          warning: (
            <div>
              Are you sure you want to remove <b>{showDeleteDialog?.name}</b>{' '}
              user from this account?
            </div>
          ),
          prompt: (
            <div>
              Type in <b>confirm</b> to continue.
            </div>
          ),
        }}
        resourceType={RESOURCE_NAME}
        show={showDeleteDialog}
        setShow={setShowDeleteDialog}
        onSubmit={async () => {
          try {
            if (!isPendingInvitation) {
              const { errors } = await api.deleteAccountMembership({
                accountName: parseName(account),
                memberId: showDeleteDialog!.id,
              });
              if (errors) {
                throw errors[0];
              }
            } else if (isPendingInvitation) {
              const { errors } = await api.deleteAccountInvitation({
                accountName: parseName(account),
                invitationId: showDeleteDialog!.id,
              });
              if (errors) {
                throw errors[0];
              }
            }
            reloadPage();
            toast.success(`${titleCase(RESOURCE_NAME)} deleted successfully`);
            setShowDeleteDialog(null);
          } catch (err) {
            handleError(err);
          }
        }}
      />
    </>
  );
};

export default UserAccessResources;
