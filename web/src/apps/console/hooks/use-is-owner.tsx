import { useCallback } from 'react';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { useConsoleApi } from '../server/gql/api-provider';

export const useIsOwner = ({ accountName }: { accountName: string }) => {
  const api = useConsoleApi();
  const { data: teamMembers, isLoading: teamMembersLoading } = useCustomSwr(
    `${accountName}-owners`,
    async () => {
      return api.listMembershipsForAccount({
        accountName,
      });
    }
  );

  const { data: currentUser, isLoading: currentUserLoading } = useCustomSwr(
    'current-user',
    async () => {
      return api.whoAmI();
    }
  );

  const isOwner = useCallback(() => {
    if (!teamMembers || !currentUser) return false;

    const owner = teamMembers.find((member) => member.role === 'account_owner');

    return owner?.user?.email === currentUser?.email;
  }, [teamMembers, currentUser]);

  return {
    isOwner: isOwner(),
    isLoading: teamMembersLoading || currentUserLoading,
  };
};
