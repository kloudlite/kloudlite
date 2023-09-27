import { Bell, UserCircle } from '@jengaicons/react';
import Popup from '~/components/molecule/popup';
import PopupSidebarLayout from '~/console/components/popup-sidebar-layout';
import { IDialog } from '~/console/components/types.d';
import ProfileAccount from '~/console/page-components/profile/profile-account';
import { UserMe } from '~/root/lib/server/gql/saved-queries';

const HandleProfile = ({ show, setShow }: IDialog<UserMe | null>) => {
  const actionItems = [
    {
      label: 'Account',
      prefix: <UserCircle />,
      value: 'account',
      panel: <ProfileAccount data={show?.data} />,
    },
    {
      label: 'Notifications',
      prefix: <Bell />,
      value: 'notifications',
      panel: <div>notification</div>,
    },
  ];
  return (
    <Popup.Root
      className="min-w-[1000px]"
      show={show as any}
      onOpenChange={(e) => setShow(e)}
    >
      <Popup.Header>Profile settings</Popup.Header>
      <Popup.Content>
        <PopupSidebarLayout items={actionItems} />
      </Popup.Content>
    </Popup.Root>
  );
};

export default HandleProfile;
