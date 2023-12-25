import { Plus } from '@jengaicons/react';
import { Outlet, useOutletContext } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import SidebarLayout from '~/console/components/sidebar-layout';

const navItems = [
  { label: 'Task', value: 'task' },
  { label: 'Task runs', value: 'taskruns' },
  { label: 'Scheduled task', value: 'scheduledtask' },
];
const CronJobs = () => {
  const rootContext: object = useOutletContext();

  return (
    <SidebarLayout
      headerActions={
        <Button variant="primary" content="Create task" prefix={<Plus />} />
      }
      navItems={navItems}
      parentPath="/jc"
      headerTitle="Jobs & Crons"
    >
      <Outlet
        context={{
          ...rootContext,
        }}
      />
    </SidebarLayout>
  );
};

export default CronJobs;
