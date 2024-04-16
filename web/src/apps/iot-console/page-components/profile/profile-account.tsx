import { Avatar } from '~/components/atoms/avatar';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import {
  BoxPrimitive,
  DeleteContainer,
} from '~/console/components/common-console-components';
import { UserMe } from '~/root/lib/server/gql/saved-queries';

const ProfileAccount = ({ data }: { data: UserMe | null | undefined }) => {
  return (
    <div className="flex flex-col gap-6xl">
      <BoxPrimitive>
        <div className="flex flex-row items-center gap-3xl">
          <Avatar size="lg" color="one" />{' '}
          <Button content="Upload photo" variant="basic" />
        </div>
        <TextInput value={data?.name} label="Full name" />
        <TextInput label="Email address" value={data?.email} disabled />
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
