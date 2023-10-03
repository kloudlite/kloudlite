import { ArrowRight, Users } from '@jengaicons/react';
import { redirect } from '@remix-run/node';
import { Link, useLoaderData, useOutletContext } from '@remix-run/react';
import { useEffect } from 'react';
import { Button } from '~/components/atoms/button';
import { usePagination } from '~/components/molecule/pagination';
import { cn, generateKey } from '~/components/utils';
import logger from '~/root/lib/client/helpers/log';
import { authBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';
import ConsoleAvatar from '../components/console-avatar';
import DynamicPagination from '../components/dynamic-pagination';
import List from '../components/list';
import RawWrapper from '../components/raw-wrapper';
import { IAccounts } from '../server/gql/queries/access-queries';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { parseName } from '../server/r-utils/common';
import { FadeIn } from './_.$account.$cluster.$project.$scope.$workspace.new-app/util';

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

    if (!accounts.length) {
      const redi: any = redirect('/new-team');
      // for tricking typescript
      return redi as {
        accounts: IAccounts;
      };
    }
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

  const { page, hasNext, hasPrevious, onNext, onPrev, setItems } =
    usePagination({
      items: accounts,
      itemsPerPage: 5,
    });

  useEffect(() => {
    setItems(accounts);
  }, [accounts]);

  return (
    <RawWrapper
      title={`Welcome ${(user?.name || '').split(' ')[0] || ''}! Select your
    Team.`}
      subtitle="Select an account to proceed to console screens."
      rightChildren={
        <FadeIn>
          <DynamicPagination
            {...{
              hasNext,
              hasPrevious,
              hasItems: accounts.length > 0,
              noItemsMessage: '0 teammates to invite.',
              onNext,
              onPrev,
              header: (
                <div className={cn('p-3xl flex flex-row text-text-default')}>
                  <div className="bodyMd">Teams for&nbsp;</div>
                  <div className="bodyMd-semibold">{email}</div>
                </div>
              ),
            }}
            className="shadow-button border border-border-default bg-surface-basic-default rounded min-h-[427px]"
          >
            <List.Root plain linkComponent={Link}>
              {page.map((account, index) => {
                const name = parseName(account);
                const displayName = account?.displayName;
                return (
                  <List.Row
                    to={`/${name}`}
                    key={name}
                    plain
                    className="group/team p-3xl [&:not(:last-child)]:border-b border-border-default last:rounded"
                    columns={[
                      {
                        key: generateKey(name, index),
                        className: 'flex-1',
                        render: () => (
                          <div className="flex flex-row items-center gap-lg">
                            <ConsoleAvatar name={name} />
                            <div className="text-text-default headingMd flex-1">
                              {displayName}{' '}
                              <span className="opacity-60">#{name}</span>
                            </div>
                          </div>
                        ),
                      },
                      {
                        key: generateKey(name, index, 'action-arrow'),
                        render: () => (
                          <div className="invisible transition-all delay-200 duration-10 group-hover/team:visible group-hover/team:translate-x-sm">
                            <ArrowRight size={24} />
                          </div>
                        ),
                      },
                    ]}
                  />
                );
              })}
            </List.Root>
          </DynamicPagination>
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
        </FadeIn>
      }
    />
  );
};

export default Accounts;
