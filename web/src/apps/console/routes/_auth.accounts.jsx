import logger from '~/root/lib/client/helpers/log';
import { useLoaderData } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { GQLServerHandler } from '../server/gql/saved-queries';

const Accounts = () => {
  const { accounts } = useLoaderData();
  return (
    <div>
      <div>Accounts</div>
      <div className="flex flex-col gap-md">
        {accounts.map(({ name }) => {
          return (
            <Button
              onClick={() => {}}
              key={name}
              content={`acocunt/${name}`}
              variant="primary-plain"
              size="medium"
            />
          );
        })}
      </div>
    </div>
  );
};

export const loader = async (ctx) => {
  let accounts;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).listAccounts(
      {}
    );
    if (errors) {
      throw errors[0];
    }
    accounts = data;
  } catch (err) {
    logger.error(err.message);
  }

  return {
    accounts: accounts || [],
  };
};

export default Accounts;
