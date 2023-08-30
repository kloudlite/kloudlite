import logger from '~/root/lib/client/helpers/log';
import { Link, useLoaderData, useOutletContext } from '@remix-run/react';
import { BrandLogo } from '~/components/branding/brand-logo';
import { Thumbnail } from '~/components/atoms/thumbnail';
import { ArrowRight, Users } from '@jengaicons/react';
import { Button } from '~/components/atoms/button';
import { cn } from '~/components/utils';
import { authBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { GQLServerHandler } from '../server/gql/saved-queries';
import RawWrapper from '../components/raw-wrapper';
import { parseDisplayname, parseName } from '../server/r-urils/common';

const Accounts = () => {
  const { accounts } = useLoaderData();
  const { user } = useOutletContext();
  const { email } = user || { email: '' };

  return (
    <RawWrapper
      leftChildren={
        <>
          <BrandLogo detailed={false} size={48} />
          <div className="flex flex-col gap-4xl">
            <div className="flex flex-col gap-3xl">
              <div className="text-text-default heading4xl">
                Welcome {(user?.name || '').split(' ')[0] || ''}! Select your
                Team.
              </div>
              <div className="text-text-default bodyMd">
                Select an account to proceed to console screens.
              </div>
            </div>
          </div>
        </>
      }
      rightChildren={
        <>
          <div className="h-7xl" />
          <div className="h-8xl" />
          <div className="flex flex-col gap-6xl">
            <div className="flex flex-col shadow-popover border border-border-default bg-surface-basic-default rounded">
              <div
                className={cn('p-3xl flex flex-row text-text-default', {
                  'border-b border-border-disabled': accounts.length,
                })}
              >
                <div className="bodyMd">Teams for&nbsp;</div>
                <div className="bodyMd-semibold">{email}</div>
              </div>
              {accounts.map((account) => {
                const name = parseName(account);
                const displayName = parseDisplayname(account);
                return (
                  <Link
                    to={`/${name}`}
                    key={name}
                    className="outline-none ring-border-focus ring-offset-1 focus:ring-2 p-3xl [&:not(:last-child)]:border-b [&:not(:last-child)]:border-border-disabled flex flex-row gap-lg items-center"
                  >
                    <Thumbnail
                      size="xs"
                      src="https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"
                    />
                    <div className="text-text-default headingMd flex-1">
                      {displayName} <span className="opacity-60">#{name}</span>
                    </div>
                    <ArrowRight size={24} />
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
    logger.error(err);
  }
  return {
    accounts: accounts || [],
  };
};

export default Accounts;
