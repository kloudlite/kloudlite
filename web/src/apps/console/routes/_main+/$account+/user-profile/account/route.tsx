import { useOutletContext } from '@remix-run/react';
import Wrapper from '~/console/components/wrapper';
import { IAccountContext } from '~/console/routes/_main+/$account+/_layout';
import ProfileAccount from '~/console/page-components/profile/profile-account';

const Account = () => {
  const { user } = useOutletContext<IAccountContext>();
  return (
    <Wrapper>
      <ProfileAccount data={user} />
    </Wrapper>
  );
};

export default Account;
