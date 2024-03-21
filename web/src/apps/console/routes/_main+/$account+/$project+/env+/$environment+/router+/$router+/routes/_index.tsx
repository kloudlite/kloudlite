import { Plus } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import Wrapper from '~/console/components/wrapper';
import { useEffect, useState } from 'react';
import { uuid } from '~/components/utils';
import { IRouter } from '~/console/server/gql/queries/router-queries';
import { NN } from '~/lib/types/common';
import HandleRoute from './handle-route';
import RouteResources from './route-resources';
import { IRouterContext } from '../_layout';
import Tools from './tools';

type ModifiedRoute = NN<IRouter['spec']['routes']>[number] & { id: string };

export type ModifiedRouter = IRouter & {
  spec: IRouter['spec'] & {
    routes?: ModifiedRoute[];
  };
};
const Router = () => {
  const { router } = useOutletContext<IRouterContext>();
  const [visible, setVisible] = useState(false);
  const [modifiedRouter, setModifiedRouter] = useState<ModifiedRouter>();

  useEffect(() => {
    const r = router.spec.routes?.map((route) => ({
      ...route,
      id: uuid(),
    }));

    const rr = router as ModifiedRouter;
    rr.spec.routes = r;
    setModifiedRouter(rr);
  }, [router]);

  return (
    <>
      <Wrapper
        header={{
          title: 'Routes',
          action: modifiedRouter &&
            (modifiedRouter?.spec?.routes || []).length > 0 && (
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
          is:
            (modifiedRouter && (modifiedRouter?.spec?.routes || []).length) ===
            0,
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
      >
        <RouteResources
          items={modifiedRouter?.spec.routes || []}
          router={modifiedRouter}
        />
      </Wrapper>
      <HandleRoute
        {...{
          visible,
          setVisible,
          isUpdate: false,
          router: modifiedRouter,
        }}
      />
    </>
  );
};

export default Router;
