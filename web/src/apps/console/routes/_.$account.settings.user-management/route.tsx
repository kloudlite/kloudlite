import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { Profile } from '~/components/molecule/profile';
import Wrapper from '~/console/components/wrapper';
import { Plus, SmileySad } from '@jengaicons/react';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useOutletContext } from '@remix-run/react';
import { useApiCall } from '~/root/lib/client/hooks/use-call-api';
import { LoadingPlaceHolder } from '~/console/components/loading';
import { NonNullableString } from '~/root/lib/types/common';
import { mapper } from '~/components/utils';
import Tools from './tools';
import HandleUser from './handle-user';
import { IAccountContext } from '../_.$account';
import Resources from './resource';

interface ITeams {
  setShowUserInvite: (fn: boolean) => void;
}

const Teams = ({ setShowUserInvite }: ITeams) => {
  const { account } = useOutletContext<IAccountContext>();
  const api = useConsoleApi();
  const { data: teamMembers, isLoading } = useApiCall(
    api.listMembershipsForAccount,
    {
      accountName: account.metadata.name,
    },
    []
  );

  return (
    <Wrapper
      empty={{
        is: teamMembers?.length === 0,
        title: 'Invite team members',
        image: <SmileySad size={48} />,
        content: <p>The Users for your teams will be listed here.</p>,
        action: {
          content: 'Invite users',
          prefix: <Plus />,
          onClick: () => {
            setShowUserInvite(true);
          },
        },
      }}
    >
      {isLoading ? (
        <LoadingPlaceHolder height={200} />
      ) : (
        <Resources
          items={mapper(teamMembers || [], (i) => ({
            id: i.user.email,
            name: i.user.name,
            role: i.role,
            email: i.user.email,
            lastLogin: '',
          }))}
        />
      )}
    </Wrapper>
  );
};

const Invitations = ({ setShowUserInvite }: ITeams) => {
  const { account } = useOutletContext<IAccountContext>();
  const api = useConsoleApi();
  const { data: invitations, isLoading } = useApiCall(
    api.listInvitationsForAccount,
    {
      accountName: account.metadata.name,
    },
    []
  );

  return (
    <Wrapper
      empty={{
        is: invitations?.filter((i) => !i.accepted)?.length === 0,
        title: 'Invite team members',
        image: <SmileySad size={48} />,
        content: (
          <p>The pending invitations for your teams will be listed here.</p>
        ),
        action: {
          content: 'Invite users',
          prefix: <Plus />,
          onClick: () => {
            setShowUserInvite(true);
          },
        },
      }}
    >
      {isLoading ? (
        <LoadingPlaceHolder height={200} />
      ) : (
        <Resources
          items={mapper(invitations?.filter((i) => !i.accepted) || [], (i) => ({
            role: i.userRole,
            name: i.userEmail || '',
            email: i.userEmail || '',
            lastLogin: '',
            id: '',
          }))}
        />
      )}
    </Wrapper>
  );
};

const SettingUserManagement = () => {
  const [active, setActive] = useState<
    'team' | 'invitations' | NonNullableString
  >('team');
  const [showUserInvite, setShowUserInvite] = useState<boolean>(false);
  return (
    <div className="flex flex-col gap-8xl">
      <div className="flex flex-col gap-3xl">
        <div className="flex flex-row gap-3xl items-center">
          <span className="flex-1 text-text-strong headingXl">
            User management
          </span>
          <Button
            content="Invite user"
            variant="primary"
            onClick={() => setShowUserInvite(true)}
          />
        </div>
        <div className="flex flex-col p-3xl gap-3xl shadow-button border border-border-default rounded bg-surface-basic-default">
          <div className="headingLg text-text-strong">Account owners</div>
          <Profile
            name="Astroman"
            subtitle="Last login was Friday, May 12, 2023 9:59 PM GMT+5:30"
            size="md"
          />

          <Profile
            name="Astroman"
            subtitle="Last login was Friday, May 12, 2023 9:59 PM GMT+5:30"
            size="md"
          />
        </div>
      </div>
      <div className="flex flex-col">
        <div className="flex flex-row gap-lgitems-center">
          <div className="flex-1">
            <ExtendedFilledTab
              value={active}
              onChange={setActive}
              items={[
                { label: 'Team member', to: 'team-member', value: 'team' },
                {
                  label: 'Pending invitation',
                  to: 'pending-invitation',
                  value: 'invitations',
                },
              ]}
            />
          </div>
          <Tools />
        </div>
        {active === 'team' ? (
          <Teams setShowUserInvite={setShowUserInvite} />
        ) : (
          <Invitations setShowUserInvite={setShowUserInvite} />
        )}
      </div>
      <HandleUser show={showUserInvite} setShow={setShowUserInvite} />
    </div>
  );
};

export default SettingUserManagement;
