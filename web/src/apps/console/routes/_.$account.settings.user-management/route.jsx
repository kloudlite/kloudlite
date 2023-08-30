import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { Profile } from '~/components/molecule/profile';
import { dummyData } from '~/console/dummy/data';
import Wrapper from '~/console/components/wrapper';
import { Plus, SmileySad } from '@jengaicons/react';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import Resource from './resource';
import Tools from './tools';
import HandleUser from './handle-user';

const SettingUserManagement = () => {
  const [active, setActive] = useState('team-member');
  const [teamMembers, _setTeamMembers] = useState(dummyData.teamMembers);
  const [showUserInvite, setShowUserInvite] = useState(null);
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
          <div className="headingLg text-text-strong">Account owner</div>
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
                { label: 'Team member', to: 'team-member' },
                { label: 'Pending invitation', to: 'pending-invitation' },
              ]}
            />
          </div>
          <Tools />
        </div>
        <Wrapper
          empty={{
            is: teamMembers.length === 0,
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
          <Resource items={teamMembers} />
        </Wrapper>
      </div>
      <HandleUser show={showUserInvite} setShow={setShowUserInvite} />
    </div>
  );
};

export default SettingUserManagement;
