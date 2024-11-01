import {
  Outlet,
  useLoaderData,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import LogoWrapper from '~/console/components/logo-wrapper';
import { BrandLogo } from '@kloudlite/design-system/branding/brand-logo';
import { Button } from '@kloudlite/design-system/atoms/button';
import { constants } from '~/console/server/utils/constants';

const Logo = () => {
  const { account } = useParams();
  const { user } = useLoaderData();
  return (
    <div className="flex flex-row items-center gap-md">
      <LogoWrapper to={`/${account}/environments`}>
        <BrandLogo />
      </LogoWrapper>
      <Button
        content={user.name}
        variant="plain"
        size="sm"
        className="!no-underline !cursor-default"
      />
    </div>
  );
};

export const handle = () => {
  return {
    logo: <Logo />,
    navbar: constants.nan,
    noBreadCrum: true,
    hideProfileName: true,
  };
};

const UserProfile = () => {
  const rootContext = useOutletContext();
  return (
    <SidebarLayout
      navItems={[
        { label: 'Account', value: 'account' },
        // { label: 'Notifications', value: 'notifications' },
        { label: 'Login connections', value: 'login-connections' },
      ]}
      parentPath="/user-profile"
      headerTitle="Profile settings"
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default UserProfile;
