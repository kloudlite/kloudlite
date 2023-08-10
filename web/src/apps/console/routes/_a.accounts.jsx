import logger from '~/root/lib/client/helpers/log';
import { Link, useLoaderData } from '@remix-run/react';
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
              key={name}
              size="md"
              variant="primary-plain"
              content={`accounts/${name}`}
              href={`/${name}/projects`}
              LinkComponent={Link}
            />
          );
        })}
      </div>
    </div>
  );
};

export const loader = async (ctx = {}) => {
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
