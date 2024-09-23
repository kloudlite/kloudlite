import { Outlet, useOutletContext } from '@remix-run/react';
import SidebarLayout from '~/console/components/sidebar-layout';
import { useHandleFromMatches } from '~/root/lib/client/hooks/use-custom-matches';

const Settings = () => {
  const rootContext = useOutletContext();
  const noLayout = useHandleFromMatches('noLayout', null);

  if (noLayout) {
    return <Outlet context={rootContext} />;
  }
  return (
    <SidebarLayout
      navItems={[
        { label: 'General', value: 'general' },
        { label: 'User management', value: 'user-management' },
        // { label: 'Cloud providers', value: 'cloud-providers' },
        { label: 'Image pull secrets', value: 'image-pull-secrets' },
        // { label: 'Image Discovery', value: 'images' },
        // { label: 'VPN', value: 'vpn' },
      ]}
      parentPath="/settings"
    // headerTitle="Settings"
    // headerActions={subNavAction.data}
    >
      <Outlet context={rootContext} />
    </SidebarLayout>
  );
};

export default Settings;
