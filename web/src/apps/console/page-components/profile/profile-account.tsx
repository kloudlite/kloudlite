import { CopySimple } from '@jengaicons/react';
import { Avatar } from '~/components/atoms/avatar';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import {
    BoxPrimitive,
    DeleteContainer,
} from '~/console/components/common-console-components';

const ProfileAccount = () => {
  return (
    <div className="flex flex-col gap-6xl">
      <BoxPrimitive>
        <div className="flex flex-row items-center gap-3xl">
          <Avatar size="lg" color="one" />{' '}
          <Button content="Upload photo" variant="basic" />
        </div>
        <TextInput value="" label="Full name" />
        <div className="flex flex-row items-center gap-3xl">
          <div className="flex-1">
            <TextInput label="Email address" value="" />
          </div>
          <div className="flex-1">
            <TextInput
              value=""
              label="Username"
              disabled
              suffixIcon={<CopySimple />}
            />
          </div>
        </div>
      </BoxPrimitive>
      <DeleteContainer title="Delete account" action={() => {}}>
        Permanently remove your personal account and all of its contents from
        the Kloudlite platform. This action is not reversible, so please
        continue with caution.
      </DeleteContainer>
    </div>
  );
};

export default ProfileAccount;
