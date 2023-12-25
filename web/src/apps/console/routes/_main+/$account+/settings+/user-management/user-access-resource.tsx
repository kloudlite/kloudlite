import { Trash } from '@jengaicons/react';
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
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';
import { parseName } from '~/console/server/r-utils/common';
import { IAccountContext } from '../../_layout';

const RESOURCE_NAME = 'user';

type BaseType = {
  id: string;
  name: string;
  role: string;
  email: string;
};

interface IResource {
  items: BaseType[];
  onDelete: (item: BaseType) => void;
}

const ExtraButton = ({ onDelete }: { onDelete: () => void }) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Remove',
          icon: <Trash size={16} />,
          type: 'item',
          onClick: onDelete,
          key: 'remove',
          className: '!text-text-critical',
        },
      ]}
    />
  );
};

const ListView = ({ items = [], onDelete }: IResource) => {
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
              render: () => <ListBody data={item.role} />,
            },
            {
              key: 3,
              render: () => <ExtraButton onDelete={() => onDelete(item)} />,
            },
          ]}
        />
      ))}
    </List.Root>
  );
};

const UserAccessResources = ({ items = [] }: { items: BaseType[] }) => {
  const [showDeleteDialog, setShowDeleteDialog] = useState<BaseType | null>(
    null
  );

  const { account } = useOutletContext<IAccountContext>();

  const api = useConsoleApi();
  const reloadPage = useReload();

  const props: IResource = {
    items,
    onDelete: (item) => {
      setShowDeleteDialog(item);
    },
  };
  return (
    <>
      <ListGridView
        listView={<ListView {...props} />}
        gridView={<ListView {...props} />}
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
            const { errors } = await api.deleteAccountMembership({
              accountName: parseName(account),
              memberId: showDeleteDialog!.id,
            });
            if (errors) {
              throw errors[0];
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
