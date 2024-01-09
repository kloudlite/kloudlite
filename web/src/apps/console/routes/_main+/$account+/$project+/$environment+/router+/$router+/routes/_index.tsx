import { Plus } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import Wrapper from '~/console/components/wrapper';
import { useState } from 'react';
import HandleRoute from './handle-route';
import RouteResources from './route-resources';
import { IRouterContext } from '../_layout';
import Tools from './tools';

const Router = () => {
  const { router } = useOutletContext<IRouterContext>();
  const [visible, setVisible] = useState(false);
  const { routes = [] } = router.spec;
  return (
    <>
      <Wrapper
        header={{
          title: 'Routes',
          action: routes.length > 0 && (
            <Button
              variant="primary"
              content="Create route"
              prefix={<Plus />}
              onClick={() => {
                setVisible(true);
              }}
            />
          ),
        }}
        empty={{
          is: routes.length === 0,
          title: 'This is where you’ll manage your Routes.',
          content: (
            <p>You can create a new routes and manage the listed routes.</p>
          ),
          action: {
            content: 'Add new route',
            prefix: <Plus />,
            onClick: () => {
              setVisible(true);
            },
          },
        }}
        tools={<Tools />}
      >
        <RouteResources items={routes} router={router} />
      </Wrapper>
      <HandleRoute
        {...{
          visible,
          setVisible,
          isUpdate: false,
          router,
        }}
      />
    </>
  );
};

export default Router;
