import SidebarLayout from '../components/sidebar-layout';

const ContainerRegistry = () => {
  return (
    <SidebarLayout
      items={[
        { label: 'General', value: 'general' },
        { label: 'Access management', value: 'access-management' },
      ]}
      parentPath="/container-registry"
      headerTitle="Container registry"
    />
  );
};

export default ContainerRegistry;
