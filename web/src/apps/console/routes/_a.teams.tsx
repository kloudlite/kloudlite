import { ArrowRight, Users } from '@jengaicons/react';
import { Link, useLoaderData, useOutletContext } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { Thumbnail } from '~/components/atoms/thumbnail';
import { cn } from '~/components/utils';
import logger from '~/root/lib/client/helpers/log';
import { authBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';
import RawWrapper from '../components/raw-wrapper';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { parseName } from '../server/r-utils/common';

export const loader = async (ctx: IRemixCtx) => {
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
    logger.error(err);
  }
  return {
    accounts: accounts || [],
  };
};

const Accounts = () => {
  const { accounts } = useLoaderData<typeof loader>();
  const { user } = useOutletContext<{
    user: UserMe;
  }>();
  const { email } = user;

  return (
    <RawWrapper
      title={`Welcome ${(user?.name || '').split(' ')[0] || ''}! Select your
    Team.`}
      subtitle="Select an account to proceed to console screens."
      rightChildren={
        <>
          <div className="h-7xl" />
          <div className="h-8xl" />
          <div className="flex flex-col gap-6xl">
            <div className="flex flex-col shadow-popover border border-border-default bg-surface-basic-default rounded">
              <div
                className={cn('p-3xl flex flex-row text-text-default', {
                  'border-b border-border-disabled': !!accounts.length,
                })}
              >
                <div className="bodyMd">Teams for&nbsp;</div>
                <div className="bodyMd-semibold">{email}</div>
              </div>
              {accounts.map((account) => {
                const name = parseName(account);
                const displayName = account?.displayName;
                return (
                  <Link
                    to={`/${name}`}
                    key={name}
                    className="group/team outline-none ring-border-focus ring-offset-1 focus:ring-2 p-3xl [&:not(:last-child)]:border-b [&:not(:last-child)]:border-border-disabled flex flex-row gap-lg items-center"
                  >
                    <Thumbnail
                      size="xs"
                      src="https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"
                    />
                    <div className="text-text-default headingMd flex-1">
                      {displayName} <span className="opacity-60">#{name}</span>
                    </div>
                    <div className="invisible transition-all group-hover/team:visible group-hover/team:translate-x-xs">
                      <ArrowRight size={24} />
                    </div>
                  </Link>
                );
              })}
            </div>
            <div className="flex flex-row gap-lg items-center py-3xl px-6xl bg-surface-basic-active rounded">
              <Users size={24} />
              <span className="text-text-default bodyMd flex-1">
                {accounts.length
                  ? 'Want to use Kloudlite with a different team?'
                  : 'Start using kloudlite by creating new team.'}
              </span>
              <Button
                variant="outline"
                content={
                  accounts.length ? 'Create another team' : 'Create new team'
                }
                LinkComponent={Link}
                to="/new-team"
              />
            </div>
            <div className="flex flex-row items-center justify-center">
              <span className="text-text-default bodyMd">
                Not able to see your team?
              </span>
              <Button
                to={`${authBaseUrl}/logout`}
                LinkComponent={Link}
                variant="primary-plain"
                content="Try a different email"
              />
            </div>
          </div>
        </>
      }
    />
  );
};

export default Accounts;
