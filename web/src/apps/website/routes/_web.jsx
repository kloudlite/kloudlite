import { BaseLayout } from '~/website/layouts/base.jsx';
import { Outlet } from '@remix-run/react';

export default () => {
  return (
    <BaseLayout>
      <Outlet />
    </BaseLayout>
  );
};
