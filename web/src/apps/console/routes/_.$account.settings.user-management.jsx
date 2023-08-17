import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Tabs from '~/components/atoms/tabs';
import { Profile } from '~/components/molecule/profile';
import List from '../components/list';

const SettingUserManagement = () => {
  const [active, setActive] = useState('team-member');
  return (
    <>
      <div className="flex flex-col gap-3xl">
        <div className="flex flex-row gap-3xl items-center">
          <span className="flex-1 text-text-strong headingXl">
            User management
          </span>
          <Button content="Invite user" variant="primary" />
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
      <div className="rounded-lg border border-border-default shadow-button flex flex-col overflow-hidden">
        <div className="flex flex-row gap-lg p-lg pr-3xl items-center bg-surface-basic-subdued border-b border-border-default">
          <div className="flex-1">
            <Tabs.Root variant="filled" value={active} onChange={setActive}>
              <Tabs.Tab
                label="Team member"
                value="team-member"
                href="team-member"
              />
              <Tabs.Tab
                label="Pending invitation"
                value="pending-invitation"
                href="pending-invitation"
              />
            </Tabs.Root>
          </div>
          <div className="flex-1">
            <TextInput placeholder="Search" value="" />
            <List />
          </div>
        </div>
      </div>
    </>
  );
};

export default SettingUserManagement;
