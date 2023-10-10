import {
  CircleFill,
  GithubLogoFill,
  ListBullets,
  LockSimple,
  LockSimpleOpen,
  Plus,
  Search,
} from '@jengaicons/react';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Pulsable from '~/components/atoms/pulsable';
import Select from '~/components/atoms/select';
import { dayjs } from '~/components/molecule/dayjs';
import { generateKey } from '~/components/utils';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { useConsoleApi } from '../server/gql/api-provider';
import List from './list';

const ADD_GITHUB_ACCOUNT_VALUE = 'add-github-account';
const SWITCH_GIT_PROVIDER_VALUE = 'switch-git-provider';

interface IGitRepoSelector {}
const GitRepoSelector = ({}: IGitRepoSelector) => {
  const api = useConsoleApi();

  const [options, setOptions] = useState<
    {
      label: string;
      labelValueIcon: JSX.Element;
      value: string;
      render: () => JSX.Element;
    }[]
  >([]);

  const { data, error, isLoading } = useCustomSwr(
    'api/github_installations',
    async () => api.listGithubInstalltions({})
  );

  const [selectedAccount, setSelectedAccount] = useState<{
    label: string;
    labelValueIcon: JSX.Element;
    value: string;
    render: () => JSX.Element;
  }>();

  useEffect(() => {
    const formattedData = data?.map((d) => ({
      label: d.account?.login || '',
      labelValueIcon: <GithubLogoFill size={14} />,
      value: `${d.id!}`,
      render: () => (
        <div className="flex flex-row gap-lg items-center">
          <div>
            <GithubLogoFill size={14} />
          </div>
          <div>{d.account?.login}</div>
        </div>
      ),
    }));

    formattedData?.push({
      label: 'Add Github Account',
      value: ADD_GITHUB_ACCOUNT_VALUE,
      labelValueIcon: <Plus size={14} />,
      render: () => (
        <div className="flex flex-row gap-lg items-center">
          <div>
            <Plus size={14} />
          </div>
          <div>Add Github Account</div>
        </div>
      ),
    });

    formattedData?.push({
      label: 'Switch Git Provider',
      value: SWITCH_GIT_PROVIDER_VALUE,
      labelValueIcon: <ListBullets size={14} />,
      render: () => (
        <div className="flex flex-row gap-lg items-center">
          <div>
            <ListBullets size={14} />
          </div>
          <div>Switch Git Provider</div>
        </div>
      ),
    });

    setOptions(formattedData || []);
    if (formattedData && formattedData?.length > 2) {
      setSelectedAccount(formattedData[0]);
    }
  }, [data]);

  const {
    data: repoData,
    error: repoError,
    isLoading: repoLoading,
  } = useCustomSwr(
    selectedAccount ? `api/github-repos-${selectedAccount?.value}` : null,
    async () =>
      api.listGithubRepos({
        installationId: parseInt(selectedAccount!.value, 10),
        pagination: {
          page: 1,
          per_page: 5,
        },
      })
  );

  return (
    <div className="p-5xl flex flex-col gap-2xl">
      <div className="heading2xl text-text-strong">Import Git Repository</div>
      <div className="flex flex-row gap-lg items-center">
        <div className="flex-1">
          <Pulsable isLoading={isLoading}>
            <Select
              options={options}
              value={selectedAccount}
              onChange={(res) => {
                if (
                  ![
                    ADD_GITHUB_ACCOUNT_VALUE,
                    SWITCH_GIT_PROVIDER_VALUE,
                  ].includes(res.value)
                ) {
                  setSelectedAccount(res);
                } else {
                }
              }}
            />
          </Pulsable>
        </div>
        <div className="flex-1">
          <Pulsable isLoading={isLoading}>
            <TextInput placeholder="Search" prefixIcon={<Search />} />
          </Pulsable>
        </div>
      </div>
      <div className="relative">
        <List.Root className="min-h-[356px]" loading={repoLoading || isLoading}>
          {repoData?.repositories?.map((repo, index) => {
            return (
              <List.Row
                key={repo.fullName}
                columns={[
                  {
                    key: generateKey(repo.fullName || '', index),
                    className: 'flex-1',
                    render: () => (
                      <div className="flex flex-row gap-lg items-center bodyMd-medium flex-1">
                        <span>{repo.fullName}</span>
                        <span>
                          {repo.private ? (
                            <LockSimple size={12} />
                          ) : (
                            <LockSimpleOpen size={12} />
                          )}
                        </span>
                        <span>
                          <CircleFill size={2} />
                        </span>
                        <span className="text-text-soft">
                          {dayjs(repo.updatedAt).fromNow()}
                        </span>
                      </div>
                    ),
                  },
                  {
                    key: generateKey(repo.fullName || '', 'action', index),
                    render: () => <Button content="Import" variant="basic" />,
                  },
                ]}
              />
            );
          })}
        </List.Root>
        <div className="absolute inset-0" />
      </div>
    </div>
  );
};

export default GitRepoSelector;
