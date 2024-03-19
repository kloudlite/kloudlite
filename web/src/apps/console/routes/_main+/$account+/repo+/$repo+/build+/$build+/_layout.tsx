import { Outlet, useOutletContext, useParams } from '@remix-run/react';
import { CommonTabs } from '~/console/components/common-navbar-tabs';
import SidebarLayout from '~/console/components/sidebar-layout';

const Tabs = () => {
  const { repo, account } = useParams();

  return (
    <CommonTabs
      backButton={{
        label: 'Builds Integrations',
        to: `/${account}/repo/${repo}/builds`,
      }}
    />
  );
};

export const handle = () => {
  return {
    navbar: <Tabs />,
  };
};

const Settings = () => {
  const rootContext = useOutletContext<any>();
  const { build } = useParams();
  return (
    <SidebarLayout
      navItems={[
        { label: 'Build runs', value: 'buildruns' },
        // { label: 'Settings', value: 'settings' },
      ]}
      parentPath={`/${build}`}
    >
      <Outlet context={{ ...rootContext }} />
    </SidebarLayout>
  );
};

export default Settings;
