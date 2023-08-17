import SidebarLayout from '../components/sidebar-layout';

const Settings = () => {
  return (
    <SidebarLayout
      items={[
        { label: 'General', value: 'general' },
        { label: 'User management', value: 'user-management' },
      ]}
      parentPath="/settings"
      headerTitle="Settings"
    />
  );
};

export default Settings;
