import { useOutletContext } from '@remix-run/react';
import { motion } from 'framer-motion';
import { useCallback, useMemo, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { dayjs } from '~/components/molecule/dayjs';
import Profile from '~/components/molecule/profile';
import { useSort } from '~/components/utils';
import { EmptyState } from '~/console/components/empty-state';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { Plus, SmileySad } from '~/console/components/icons';
import SecondarySubHeader from '~/console/components/secondary-sub-header';
import Wrapper from '~/console/components/wrapper';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName } from '~/console/server/r-utils/common';
import Pulsable from '~/root/lib/client/components/pulsable';
import { useSearch } from '~/root/lib/client/helpers/search-filter';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { ExtractArrayType, NonNullableString } from '~/root/lib/types/common';
import { IAccountContext } from '../../_layout';
import { ISettingsContext } from '../_layout';
import HandleUser from './handle-user';
import Tools from './tools';
import UserAccessResources from './user-access-resource';

interface ITeams {
  setShowUserInvite: React.Dispatch<React.SetStateAction<boolean>>;
  searchText: string;
  sortTeamMembers: {
    sortByProperty: string;
    sortByTime: string;
  };
  isOwner?: boolean;
}

const placeHolderUsers = Array(3)
  .fill(0)
  .map((_, i) => ({
    id: `${i}`,
    name: 'sample user',
    role: 'account_owner',
    email: 'sampleuser@gmail.com',
  }));

const Teams = ({
  setShowUserInvite,
  searchText,
  sortTeamMembers,
  isOwner,
}: ITeams) => {
  const { account } = useOutletContext<IAccountContext>();
  const api = useConsoleApi();
  const { data: teamMembers, isLoading } = useCustomSwr(
    `${parseName(account)}-teams`,
    async () => {
      return api.listMembershipsForAccount({
        accountName: parseName(account),
      });
    }
  );

  const searchResp = useSearch(
    {
      data:
        teamMembers?.map((i) => {
          return {
            ...i,
            searchField: i.user.name,
          };
        }) || [],
      searchText,
      keys: ['searchField'],
    },
    [searchText, teamMembers]
  );

  const sortFunction = useCallback(
    ({
      a,
      b,
    }: {
      a: ExtractArrayType<typeof teamMembers>;
      b: ExtractArrayType<typeof teamMembers>;
    }) => {
      const isAscending = sortTeamMembers.sortByTime === 'asc';

      const x = isAscending ? a : b;
      const y = isAscending ? b : a;

      if (sortTeamMembers.sortByProperty === 'name') {
        return x.user.name.localeCompare(y.user.name);
      }

      return dayjs(x.user.joined).unix() - dayjs(y.user.joined).unix();
    },
    [sortTeamMembers]
  );

  const sorted = useSort(searchResp || [], (a, b) => sortFunction({ a, b }));

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ ease: 'anticipate', duration: 0.1 }}
    >
      <Wrapper
        empty={{
          is: teamMembers?.length === 0,
          title: 'Invite team members',
          image: <SmileySad size={48} />,
          content: <p>The Users for your teams will be listed here.</p>,
          action: {
            content: 'Invite users',
            prefix: <Plus />,
            onClick: () => {
              setShowUserInvite(true);
            },
          },
        }}
      >
        <Pulsable isLoading={isLoading}>
          {!isLoading && searchResp.length === 0 && (
            <EmptyState
              {...{
                image: <SmileySad size={48} />,
                heading: 'No users found',
              }}
            />
          )}

          {(isLoading || searchResp.length > 0) && (
            <UserAccessResources
              items={
                isLoading && searchResp.length === 0
                  ? placeHolderUsers
                  : sorted.map((i) => ({
                      id: i.user.id,
                      name: i.user.name,
                      role: i.role,
                      email: i.user.email,
                    }))
              }
              isPendingInvitation={false}
              isOwner={isOwner || false}
            />
          )}
        </Pulsable>
      </Wrapper>
    </motion.div>
  );
};

const Invitations = ({
  setShowUserInvite,
  searchText,
  sortTeamMembers,
  isOwner,
}: ITeams) => {
  const { account } = useOutletContext<IAccountContext>();
  const api = useConsoleApi();

  const { data: invitations, isLoading } = useCustomSwr(
    `${parseName(account)}-invitations`,
    async () => {
      return api.listInvitationsForAccount({
        accountName: parseName(account),
      });
    }
  );

  const searchResp = useSearch(
    {
      data:
        invitations?.map((i) => {
          return {
            ...i,
            searchField: i.userEmail,
          };
        }) || [],
      searchText,
      keys: ['searchField'],
    },
    [searchText, invitations]
  );

  const sortFunction = useCallback(
    ({
      a,
      b,
    }: {
      a: ExtractArrayType<typeof invitations>;
      b: ExtractArrayType<typeof invitations>;
    }) => {
      const isAscending = sortTeamMembers.sortByTime === 'asc';

      const x = isAscending ? a : b;
      const y = isAscending ? b : a;

      // TODO: remove below if
      if (!x.userEmail || !y.userEmail) {
        return 0;
      }

      if (sortTeamMembers.sortByProperty === 'name') {
        return x.userEmail.localeCompare(y.userEmail);
      }

      return dayjs(x.creationTime).unix() - dayjs(y.creationTime).unix();
    },
    [sortTeamMembers]
  );

  const sorted = useSort(searchResp || [], (a, b) => sortFunction({ a, b }));

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ ease: 'anticipate', duration: 0.1 }}
    >
      <Wrapper
        empty={{
          is: invitations?.filter((i) => !i.accepted)?.length === 0,
          title: 'Invite team members',
          image: <SmileySad size={48} />,
          content: (
            <p>The pending invitations for your teams will be listed here.</p>
          ),
          action: {
            content: 'Invite users',
            prefix: <Plus />,
            onClick: () => {
              setShowUserInvite(true);
            },
          },
        }}
      >
        <Pulsable isLoading={isLoading}>
          <UserAccessResources
            items={
              isLoading && sorted.length === 0
                ? placeHolderUsers
                : searchResp
                    ?.filter((i) => !i.accepted)
                    .map((i) => ({
                      role: i.userRole,
                      name: i.userEmail || '',
                      email: i.userEmail || '',
                      id: i.id || i.userEmail || '',
                    }))
            }
            isPendingInvitation
            isOwner={isOwner || false}
          />
        </Pulsable>
      </Wrapper>
    </motion.div>
  );
};

const SettingUserManagement = () => {
  const [active, setActive] = useState<
    'team' | 'invitations' | NonNullableString
  >('team');
  const [visible, setVisible] = useState(false);
  // const { account } = useOutletContext<IAccountContext>();
  const { teamMembers, currentUser } = useOutletContext<ISettingsContext>();

  const [searchText, setSearchText] = useState('');

  const [sortByProperty, setSortbyProperty] = useState({
    sortByProperty: 'updated',
    sortByTime: 'des',
  });

  // const api = useConsoleApi();

  const isOwner = useMemo(() => {
    if (!teamMembers || !currentUser) return false;
    const owner = teamMembers.find((member) => member.role === 'account_owner');
    return owner?.user?.email === currentUser?.email;
  }, [teamMembers, currentUser]);

  // const { data: teamMembers, isLoading } = useCustomSwr(
  //   `${parseName(account)}-owners`,
  //   async () => {
  //     return api.listMembershipsForAccount({
  //       accountName: parseName(account),
  //     });
  //   }
  // );

  // const owners = useCallback(
  //   () => teamMembers?.filter((i) => i.role === 'account_owner') || [],
  //   [teamMembers]
  // )();

  const accountOwner = teamMembers?.find((i) => i.role === 'account_owner');

  return (
    <div className="flex flex-col gap-8xl">
      <div className="flex flex-col gap-6xl">
        <SecondarySubHeader
          title="User management"
          action={
            isOwner && (
              <Button
                content="Invite user"
                variant="primary"
                onClick={() => setVisible(true)}
              />
            )
          }
        />

        <div className="flex flex-col p-3xl gap-3xl shadow-button border border-border-default rounded bg-surface-basic-default">
          <div className="headingLg text-text-strong">Account owners</div>

          <Pulsable isLoading={false}>
            <div className="flex flex-col gap-3xl">
              <Profile
                key={accountOwner?.user?.email}
                name={accountOwner?.user?.name}
                subtitle={accountOwner?.user?.email}
              />
            </div>
          </Pulsable>

          {/* <Pulsable isLoading={isLoading}>
            <div className="flex flex-col gap-3xl">
              {[
                ...(isLoading
                  ? [
                      {
                        user: {
                          email: 'sampleuser@gmail.com',
                          name: 'sample user',
                        },
                      },
                    ]
                  : owners),
              ].map((t) => {
                return (
                  <Profile
                    key={t.user.email}
                    name={t.user.name}
                    subtitle={t.user.email}
                  />
                );
              })}
            </div>
          </Pulsable> */}
        </div>
      </div>
      <div className="flex flex-col">
        <div className="flex flex-row gap-lg items-center pb-3xl">
          <div className="flex-1">
            <ExtendedFilledTab
              value={active}
              onChange={setActive}
              items={[
                { label: 'Team members', to: 'team-member', value: 'team' },
                {
                  label: 'Pending invitations',
                  to: 'pending-invitation',
                  value: 'invitations',
                },
              ]}
            />
          </div>
          <Tools
            setSearchText={setSearchText}
            searchText={searchText}
            sortTeamMembers={setSortbyProperty}
          />
        </div>
        {active === 'team' ? (
          <Teams
            setShowUserInvite={setVisible}
            searchText={searchText}
            sortTeamMembers={sortByProperty}
            isOwner={isOwner}
          />
        ) : (
          <Invitations
            setShowUserInvite={setVisible}
            searchText={searchText}
            sortTeamMembers={sortByProperty}
            isOwner={isOwner}
          />
        )}
      </div>
      <HandleUser {...{ isUpdate: false, visible, setVisible }} />
    </div>
  );
};

export default SettingUserManagement;
