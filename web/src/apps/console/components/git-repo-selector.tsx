import {
  CircleFill,
  GithubLogoFill,
  GitlabLogoFill,
  ListBullets,
  LockSimple,
  LockSimpleOpen,
  Plus,
  Search,
} from '@jengaicons/react';
import { AnimatePresence, motion } from 'framer-motion';
import { useCallback, useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { dayjs } from '~/components/molecule/dayjs';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { githubAppName } from '~/root/lib/configs/base-url.cjs';
import { SWRResponse } from 'swr';
import Radio from '~/components/atoms/radio';
import { useConsoleApi } from '../server/gql/api-provider';
import {
  IGithubInstallations,
  IGitlabGroups,
} from '../server/gql/queries/git-queries';
import { popupWindow } from '../utils/commons';
import Pulsable from './pulsable';

const ADD_GITHUB_ACCOUNT_VALUE = 'add-github-account';
const SWITCH_GIT_PROVIDER_VALUE = 'switch-git-provider';

const useDebounceText = (value: string, delay: number) => {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
};

const useGithubProvider = ({
  fetchInstallations = true,
}: {
  fetchInstallations?: boolean;
}) => {
  const [organization, setOrganization] = useState<string | null>(null);
  const [searchText, setSearchText] = useState('');
  const [repoUrl, setRepoUrl] = useState<string | null>(null);
  const [_fetchInstallations, setFetchInstallations] =
    useState(fetchInstallations);

  const api = useConsoleApi();
  const data = {
    key: 'api/github-installations',
    repoKey: 'api/github-repos',
    branchKey: 'api/github-branches',
    api: api.listGithubInstalltions,
    repoApi: api.searchGithubRepos,
    branchApi: api.listGithubBranches,
  };

  const debouncedSearch = useDebounceText(searchText, 500);
  return {
    installations: useCustomSwr(
      () => (_fetchInstallations ? data.key : null),
      async () => data.api({})
    ),
    repos: useCustomSwr(
      () =>
        organization
          ? `${data.repoKey}-${organization}-${debouncedSearch}`
          : null,
      async () =>
        data.repoApi({
          organization: organization || '',
          search: searchText,
          pagination: {
            page: 1,
            per_page: 5,
          },
        })
    ),
    branches: useCustomSwr(
      () => (repoUrl ? `${data.branchKey}-${repoUrl}` : null),
      async () =>
        data.branchApi({
          repoUrl: repoUrl!,
        })
    ),
    setOrganization,
    setSearchText,
    setRepoUrl,
    setFetchInstallations,
  };
};

const useGitlabProvider = ({
  fetchGroups = true,
}: {
  fetchGroups?: boolean;
}) => {
  const [searchText, setSearchText] = useState('');
  const [groupId, setGroupId] = useState<string | null>(null);
  const [repoId, setRepId] = useState<string | null>(null);
  const [_fetchGroups, setFetchGroups] = useState(fetchGroups);

  const api = useConsoleApi();
  const data = {
    key: 'api/gitlab-installations',
    repoKey: 'api/gitlab-repos',
    branchKey: 'api/gitlab-branches',
    api: api.listGitlabGroups,
    repoApi: api.listGitlabRepos,
    branchApi: api.listGitlabBranches,
  };
  const debouncedSearch = useDebounceText(searchText, 500);

  return {
    groups: useCustomSwr(
      () => (_fetchGroups ? data.key : null),
      async () => data.api({})
    ),
    repos: useCustomSwr(
      () => (groupId ? `${data.repoKey}-${groupId}-${debouncedSearch}` : null),
      async () =>
        data.repoApi({
          groupId: groupId!,
          query: searchText,
          pagination: {
            page: 1,
            per_page: 5,
          },
        })
    ),
    branches: useCustomSwr(
      () => (repoId ? `${data.branchKey}-${repoId}` : null),
      async () =>
        data.branchApi({
          repoId: repoId!,
        })
    ),
    setGroupId,
    setRepId,
    setFetchGroups,
    setSearchText,
  };
};
interface IBranch {
  repo: string;
  branch: string;
  provider: 'github' | 'gitlab';
}

interface IListRenderer {
  response: SWRResponse<
    { name: string; updatedAt: any; private: true; url: string }[],
    any,
    any
  >;
  onChange: (value: string) => void;
  value: string;
}

const mockListData = [
  {
    name: 'this is moc name',
    private: false,
    updatedAt: 'a month ago',
    url: '',
  },
  {
    name: 'this is moc name1',
    private: false,
    updatedAt: 'a month ago',
    url: '',
  },
  {
    name: 'this is moc name2',
    private: false,
    updatedAt: 'a month ago',
    url: '',
  },
  {
    name: 'this is moc name3',
    private: false,
    updatedAt: 'a month ago',
    url: '',
  },
  {
    name: 'this is moc name4',
    private: false,
    updatedAt: 'a month ago',
    url: '',
  },
];

const ListRenderer = ({ response, onChange, value }: IListRenderer) => {
  const { isLoading, data } = response;

  return (
    <div className="flex flex-col">
      <Pulsable isLoading={isLoading || !data}>
        <Radio.Root
          value={value}
          withBounceEffect={false}
          labelPlacement="left"
          className="!gap-0"
          onChange={onChange}
        >
          {[...(isLoading || !data ? mockListData : data)].map((repo) => {
            return (
              <Radio.Item
                key={repo.name}
                value={repo.url}
                className="pulsable flex-row justify-between w-full py-2xl bodyMd-medium"
                labelPlacement="left"
              >
                <div className="flex flex-row items-center gap-lg bodyMd-medium">
                  <span>{repo.name}</span>

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
              </Radio.Item>
            );
          })}
        </Radio.Root>
      </Pulsable>
    </div>
  );
};

interface IRepoRender {
  label: string;
  labelValueIcon: JSX.Element;
  value: string;
  render: () => JSX.Element;
}

type IFormattedData =
  | ({ data: IGithubInstallations } & { provider: 'github' })
  | ({ data: IGitlabGroups } & { provider: 'gitlab' });

const formatData = (data: IFormattedData) => {
  let formattedData: IRepoRender[] = [];

  const iconSize = 16;

  switch (data.provider) {
    case 'github':
      formattedData = data.data?.map((d) => ({
        label: d.account?.login || '',
        labelValueIcon: <GithubLogoFill size={iconSize} />,
        value: `${d.id!}`,
        render: () => (
          <div className="flex flex-row gap-lg items-center">
            <div>
              <GithubLogoFill size={iconSize} />
            </div>
            <div>{d.account?.login}</div>
          </div>
        ),
      }));

      formattedData?.push({
        label: 'Add Github Account',
        value: ADD_GITHUB_ACCOUNT_VALUE,
        labelValueIcon: <Plus size={iconSize} />,
        render: () => (
          <div className="flex flex-row gap-lg items-center">
            <div>
              <Plus size={iconSize} />
            </div>
            <div>Add Github Account</div>
          </div>
        ),
      });

      break;
    case 'gitlab':
      formattedData = data.data?.map((d) => ({
        label: d.fullName || '',
        labelValueIcon: <GitlabLogoFill size={iconSize} />,
        value: `${d.id!}`,
        render: () => (
          <div className="flex flex-row gap-lg items-center">
            <div>
              <GitlabLogoFill size={iconSize} />
            </div>
            <div>{d.fullName}</div>
          </div>
        ),
      }));
      break;
    default:
      formattedData = [];
      break;
  }

  formattedData?.push({
    label: 'Switch Git Provider',
    value: SWITCH_GIT_PROVIDER_VALUE,
    labelValueIcon: <ListBullets size={iconSize} />,
    render: () => (
      <div className="flex flex-row gap-lg items-center">
        <div>
          <ListBullets size={iconSize} />
        </div>
        <div>Switch Git Provider</div>
      </div>
    ),
  });

  return formattedData;
};

interface IGitRepoSelector {
  onChange?(source: IBranch): void;
  error?: string;
  value?: {
    provider: 'github' | 'gitlab';
    repository: string;
    branch: string;
  };
}
const GitRepoSelector = ({ onChange, error, value }: IGitRepoSelector) => {
  const githubInstallUrl = `https://github.com/apps/${githubAppName}/installations/new`;

  const [showProviderSwitch, setProviderSwitch] = useState(false);

  const [searchText, setSearchText] = useState('');

  const [provider, setProvider] = useState<'github' | 'gitlab'>('github');
  const [repo, setRepo] = useState('');
  const {
    installations,
    repos: githubRepos,
    setOrganization,
    setSearchText: setGithubRepoSearchText,
    setFetchInstallations,
    branches: githubBranches,
    setRepoUrl,
  } = useGithubProvider({});

  const {
    groups,
    repos: gitlabRepos,
    setFetchGroups,
    setGroupId,
    setRepId,
    branches: gitlabBranches,
    setSearchText: setGitlabRepoSearchText,
  } = useGitlabProvider({
    fetchGroups: false,
  });

  const [selectedBranch, setSelectedBranch] = useState<IBranch | null>();
  const [options, setOptions] = useState<IRepoRender[]>([]);

  const [selectedAccount, setSelectedAccount] = useState<IRepoRender>();

  const setupRepo = useCallback((v: string) => {
    setRepo(v);
    if (value) {
      setSelectedBranch({
        branch: value.branch,
        provider: value.provider,
        repo: value.repository,
      });
      onChange?.({ branch: value.branch, provider, repo: value.repository });
    } else {
      setSelectedBranch(null);
      onChange?.({ branch: '', provider, repo: 'null' });
    }

    if (provider === 'github') {
      setRepoUrl(v);
    } else {
      setRepId(v);
    }
  }, []);

  useEffect(() => {
    if (value) {
      setProvider(value.provider);
    }
    if (!repo) {
      switch (provider) {
        case 'github':
          setupRepo(
            value ? value.repository : githubRepos.data?.[0]?.url || ''
          );
          break;
        case 'gitlab':
          setupRepo(
            value ? value.repository : gitlabRepos.data?.[0]?.url || ''
          );
          break;
        default:
          break;
      }
    }
  }, [githubRepos.data, gitlabRepos.data, value]);

  useEffect(() => {
    let formattedData: IRepoRender[] = [];

    if (installations?.data && provider === 'github') {
      setOrganization(
        installations.data?.[0]?.account?.login
          ? `${installations.data?.[0].account?.login}`
          : null
      );

      formattedData = formatData({
        data: installations.data,
        provider,
      });
    }

    if (groups?.data && provider === 'gitlab') {
      setGroupId(groups.data?.[0]?.id);
      formattedData = formatData({ data: groups.data, provider });
    }

    setOptions(formattedData || []);
    if (
      formattedData &&
      ((provider === 'github' && formattedData?.length > 2) ||
        (provider === 'gitlab' && formattedData?.length > 1))
    ) {
      setSelectedAccount(formattedData[0]);
    }
  }, [installations.data, groups.data]);

  useEffect(() => {
    if (showProviderSwitch) {
      setOrganization(null);
      setGroupId(null);
      setFetchGroups(false);
      setFetchInstallations(false);
      setSelectedBranch(null);
      onChange?.({ branch: '', provider, repo: 'null' });
    }
  }, [showProviderSwitch]);

  useEffect(() => {
    switch (provider) {
      case 'github':
        setGithubRepoSearchText(searchText);
        break;
      case 'gitlab':
        setGitlabRepoSearchText(searchText);
        break;
      default:
        break;
    }
  }, [searchText]);

  const valueRender = ({ label, labelValueIcon }: IRepoRender) => {
    return (
      <div className="flex flex-row gap-xl items-center bodyMd text-text-default">
        <span>{labelValueIcon}</span>
        <span>{label}</span>
      </div>
    );
  };

  return (
    <div>
      <div className="flex flex-col relative border border-border-default rounded px-2xl">
        <div className="flex flex-row gap-lg items-center py-2xl">
          <div className="flex-1">
            <Pulsable isLoading={installations.isLoading || groups.isLoading}>
              <div className="pulsable">
                <Select
                  size="lg"
                  valueRender={valueRender}
                  disabled={showProviderSwitch}
                  options={async () => options}
                  value={selectedAccount}
                  onChange={(res) => {
                    if (
                      ![
                        ADD_GITHUB_ACCOUNT_VALUE,
                        SWITCH_GIT_PROVIDER_VALUE,
                      ].includes(res.value)
                    ) {
                      setRepo('');
                      setSelectedAccount(res);
                      switch (provider) {
                        case 'github':
                          setOrganization(res.label);
                          break;
                        case 'gitlab':
                          setGroupId(res.value);
                          break;
                        default:
                          break;
                      }
                    } else if (res.value === SWITCH_GIT_PROVIDER_VALUE) {
                      setProviderSwitch(true);
                    } else if (res.value === ADD_GITHUB_ACCOUNT_VALUE) {
                      popupWindow({
                        url: githubInstallUrl,
                      });
                    }
                  }}
                />
              </div>
            </Pulsable>
          </div>
          <div className="flex-1">
            <Pulsable isLoading={installations.isLoading || groups.isLoading}>
              <TextInput
                size="lg"
                placeholder="Search"
                prefixIcon={<Search />}
                value={searchText}
                onChange={({ target }) => {
                  setSearchText(target.value);
                }}
              />
            </Pulsable>
          </div>
        </div>
        <ListRenderer
          value={repo}
          response={provider === 'github' ? githubRepos : gitlabRepos}
          onChange={setupRepo}
        />
        <AnimatePresence mode="wait">
          {showProviderSwitch && (
            <motion.div className="absolute z-10 inset-0 flex flex-col items-center justify-center bg-surface-basic-subdued border border-border-default rounded">
              <div className="text-text-soft bodyMd mb-5xl">
                Select a Git provider to import an existing project from a Git
                Repository.
              </div>
              <div className="w-[320px] flex flex-col gap-lg">
                <Button
                  variant="tertiary"
                  content="Continue with Github"
                  prefix={<GithubLogoFill />}
                  block
                  onClick={() => {
                    setProvider('github');
                    setFetchInstallations(true);
                    setFetchGroups(false);
                    setProviderSwitch(false);
                  }}
                />
                <Button
                  variant="purple"
                  content="Continue with Gitlab"
                  prefix={<GitlabLogoFill />}
                  block
                  onClick={() => {
                    setProvider('gitlab');
                    setFetchInstallations(false);
                    setFetchGroups(true);
                    setProviderSwitch(false);
                  }}
                />
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
      <div className="pt-3xl">
        <Pulsable isLoading={installations.isLoading || groups.isLoading}>
          <div className="pulsable">
            <Select
              label="Branch"
              size="lg"
              value={
                selectedBranch
                  ? {
                      label: selectedBranch.branch,
                      value: selectedBranch.branch,
                    }
                  : undefined
              }
              loading={githubBranches.isLoading || gitlabBranches.isLoading}
              disableWhileLoading
              disabled={!repo}
              placeholder="Select a branch"
              options={async () =>
                (
                  (provider === 'github' ? githubBranches : gitlabBranches)
                    .data || []
                )?.map((d) => ({ label: d.name || '', value: d.name || '' }))
              }
              onChange={({ value }) => {
                setSelectedBranch({ branch: value, provider, repo });
                onChange?.({ branch: value, provider, repo });
              }}
              error={!!error}
              message={error}
            />
          </div>
        </Pulsable>
      </div>
    </div>
  );
};

export default GitRepoSelector;
