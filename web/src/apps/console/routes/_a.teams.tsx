import { ArrowRight, Users } from '@jengaicons/react';
import { redirect } from '@remix-run/node';
import {
  Link,
  useLoaderData,
  useNavigate,
  useOutletContext,
} from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { usePagination } from '~/components/molecule/pagination';
import { cn, generateKey } from '~/components/utils';
import logger from '~/root/lib/client/helpers/log';
import { authBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { UserMe } from '~/root/lib/server/gql/saved-queries';
import { IRemixCtx } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import ConsoleAvatar from '../components/console-avatar';
import DynamicPagination from '../components/dynamic-pagination';
import List from '../components/list';
import RawWrapper from '../components/raw-wrapper';
import { IAccounts, IInvites } from '../server/gql/queries/access-queries';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { parseName } from '../server/r-utils/common';
import { FadeIn } from './_.$account.$cluster.$project.$scope.$workspace.new-app/util';
import { useConsoleApi } from '../server/gql/api-provider';

export const loader = async (ctx: IRemixCtx) => {
  let accounts;
  let invites;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).listAccounts(
      {}
    );
    if (errors) {
      throw errors[0];
    }
    accounts = data;

    const { data: dataInvite, errors: errorsInvite } = await GQLServerHandler(
      ctx.request
    ).listInvitationsForUser({ onlyPending: true });

    if (errorsInvite) {
      throw errorsInvite[0];
    }

    invites = dataInvite;

    if (!(accounts.length || invites.length)) {
      const redi: any = redirect('/new-team');
      // for tricking typescript
      return redi as {
        accounts: IAccounts;
        invites: IInvites;
      };
    }
  } catch (err) {
    logger.error(err);
  }
  return {
    accounts: accounts || [],
    invites: invites || [],
  };
};

const Accounts = () => {
  const { accounts, invites } = useLoaderData<typeof loader>();
  const { user } = useOutletContext<{
    user: UserMe;
  }>();

  const [isHandling, setIsHandling] = useState<
    'none' | 'accepting' | 'rejecting'
  >('none');

  const { email } = user;

  const formatData = () => {
    return [
      ...invites.map((invite) => ({
        id: invite.id,
        updateTime: invite.updateTime,
        displayName: invite.accountName,
        metadata: {
          name: invite.accountName,
          annotations: invite.accountName,
        },
        isInvite: true,
        inviteToken: invite.inviteToken,
      })),
      ...accounts.map((account) => ({
        ...account,
        isInvite: false,
        inviteToken: null,
      })),
    ];
  };
  const { page, hasNext, hasPrevious, onNext, onPrev, setItems } =
    usePagination({
      items: formatData(),
      itemsPerPage: 5,
    });

  useEffect(() => {
    setItems(formatData());
  }, [accounts]);

  const api = useConsoleApi();
  const navigate = useNavigate();
  const reload = useReload();

  const handleInvitation = async ({
    accountName,
    inviteToken,
    api,
    success,
  }: {
    accountName: string;
    inviteToken: string;
    api: any;
    success: () => void;
  }) => {
    try {
      const { errors } = await api({
        accountName,
        inviteToken,
      });

      if (errors) {
        throw errors[0];
      }
      success();
    } catch (err) {
      handleError(err);
    } finally {
      setIsHandling('none');
    }
  };

  const acceptInvitation = async ({
    accountName,
    inviteToken,
  }: {
    accountName: string;
    inviteToken: string;
  }) => {
    setIsHandling('accepting');
    await handleInvitation({
      accountName,
      inviteToken,
      api: api.acceptInvitation,
      success: () => {
        navigate(`/${accountName}/infra/clusters`);
      },
    });
  };

  const rejectInvitation = async ({
    accountName,
    inviteToken,
  }: {
    accountName: string;
    inviteToken: string;
  }) => {
    setIsHandling('rejecting');

    await handleInvitation({
      accountName,
      inviteToken,
      api: api.rejectInvitation,
      success: () => {
        reload();
      },
    });
  };

  return (
    <RawWrapper
      title={`Welcome ${(user?.name || '').split(' ')[0] || ''}! Select your
    Team.`}
      subtitle="Select an account to proceed to console screens."
      rightChildren={
        <FadeIn notForm>
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
                const { isInvite, displayName, inviteToken } = account;
                return (
                  <List.Row
                    {...(isInvite ? {} : { to: `/${name}/infra` })}
                    key={name}
                    plain
                    className={cn(
                      'group/team p-3xl [&:not(:last-child)]:border-b border-border-disabled last:rounded',
                      {
                        '!cursor-default': isInvite,
                      }
                    )}
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
                        ...(isInvite
                          ? {
                              key: generateKey(name, index, 'action-arrow'),
                              render: () => (
                                <div className="flex flex-row gap-lg items-center">
                                  <Button
                                    content="Decline"
                                    size="sm"
                                    variant="basic"
                                    onClick={() => {
                                      rejectInvitation({
                                        accountName: displayName,
                                        inviteToken: inviteToken || '',
                                      });
                                    }}
                                    loading={isHandling === 'rejecting'}
                                  />
                                  <Button
                                    content="Accept invitation"
                                    size="sm"
                                    onClick={() =>
                                      acceptInvitation({
                                        accountName: displayName,
                                        inviteToken: inviteToken || '',
                                      })
                                    }
                                    loading={isHandling === 'accepting'}
                                  />
                                </div>
                              ),
                            }
                          : {
                              key: generateKey(name, index, 'action-arrow'),
                              render: () => (
                                <div className="invisible transition-all delay-100 duration-10 group-hover/team:visible group-hover/team:translate-x-sm">
                                  <ArrowRight size={24} />
                                </div>
                              ),
                            }),
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
