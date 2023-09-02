import { Plus } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { Link, useLoaderData, useOutletContext } from '@remix-run/react';
import AlertDialog from '~/console/components/alert-dialog';
import Wrapper from '~/console/components/wrapper';
import logger from '~/root/lib/client/helpers/log';
import { getPagination, getSearch } from '~/console/server/r-urils/common';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import {
  ensureAccountSet,
  ensureClusterSet,
} from '~/console/server/utils/auth-utils';
import { parseError } from '~/root/lib/utils/common';
import SecretResource from '~/console/page-components/secret-resource';
import { parseNodes } from '~/console/server/utils/kresources/aggregated';
import Tools from './tools';
import HandleSecret from './handle-secret';

const Secrets = () => {
  const [showHandleSecret, setHandleSecret] = useState(null);
  const [showDeleteSecret, setShowDeleteSecret] = useState(false);

  const data = useOutletContext();

  useEffect(() => {
    if (data?.setSubNavAction) {
      data.setSubNavAction({
        action: () => {
          setHandleSecret({ type: 'add', data: null });
        },
      });
    }
  }, []);

  const { promise } = useLoaderData();

  return (
    <>
      <LoadingComp data={promise}>
        {({ secretsData }) => {
          const secrets = parseNodes(secretsData);
          if (!secrets) {
            return null;
          }
          return (
            <Wrapper
              empty={{
                is: secrets.length === 0,
                title: 'This is where you’ll manage your Secret.',
                content: (
                  <p>
                    You can create a new secret and manage the listed secrets.
                  </p>
                ),
                action: {
                  content: 'Create secret',
                  prefix: <Plus />,
                  LinkComponent: Link,
                  onClick: () => {
                    setHandleSecret({ type: 'add', data: null });
                  },
                },
              }}
            >
              <Tools />
              <SecretResource items={secrets} linkComponent={Link} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleSecret show={showHandleSecret} setShow={setHandleSecret} />
      {/* Alert Dialog for deleting secret */}
      <AlertDialog
        show={showDeleteSecret}
        setShow={setShowDeleteSecret}
        title="Delete secret"
        message={"Are you sure you want to delete 'kloud-root-ca.crt"}
        type="critical"
        okText="Delete"
        onSubmit={() => {}}
      />
    </>
  );
};

export default Secrets;

export const handle = {
  subheaderAction: () => <Button content="Add new secret" prefix={<Plus />} />,
};

export const loader = async (ctx) => {
  ensureAccountSet(ctx);
  ensureClusterSet(ctx);
  const { project, scope, workspace } = ctx.params;

  const promise = pWrapper(async () => {
    try {
      const { data, errors } = await GQLServerHandler(ctx.request).listSecrets({
        project: {
          value: project,
          type: 'name',
        },
        scope: {
          value: workspace,
          type: scope === 'workspace' ? 'workspaceName' : 'environmentName',
        },
        pagination: getPagination(ctx),
        search: getSearch(ctx),
      });
      if (errors) {
        throw errors[0];
      }
      return { secretsData: data };
    } catch (err) {
      logger.error(err);
      return { error: parseError(err).message };
    }
  });

  return defer({ promise });
};
