import {
  CircleFill,
  GitBranch,
  GithubLogoFill,
  GitlabLogoFill,
  ListBullets,
  LockSimple,
  LockSimpleOpen,
  Pencil,
  Plus,
  Search,
} from '@jengaicons/react';
import { AnimatePresence, motion } from 'framer-motion';
import { useEffect, useState } from 'react';
import { Button } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { dayjs } from '~/components/molecule/dayjs';
import { generateKey } from '~/components/utils';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import { githubAppName } from '~/root/lib/configs/base-url.cjs';
import Yup from '~/root/lib/server/helpers/yup';
import { SWRResponse } from 'swr';
import ButtonGroup from '~/components/atoms/button-group';
import { useConsoleApi } from '../server/gql/api-provider';
import {
  IGithubInstallations,
  IGitlabGroups,
} from '../server/gql/queries/git-queries';
import { DIALOG_TYPE, popupWindow } from '../utils/commons';
import List from './list';
import Pulsable from './pulsable';
import { IDialog, IShowDialog } from './types.d';

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
  branch: string | null | undefined;
  provider: 'github' | 'gitlab';
}

const BranchChooser = ({
  show,
  setShow,
  onSubmit,
}: IDialog<IBranch | null, IBranch>) => {
  const [fetchedBranches, setFetchedBranches] = useState<
    { label: string; value: string }[]
  >([]);
  const { setRepoUrl: setGithubRepoUrl, branches: githubBrances } =
    useGithubProvider({
      fetchInstallations: false,
    });

  const { setRepId: setGitlabRepoUrl, branches: gitlabBrances } =
    useGitlabProvider({ fetchGroups: false });

  const [selectedBranch, setSelectedBranch] = useState<
    (typeof fetchedBranches)[number] | undefined
  >();

  const onClose = (e: any) => {
    setSelectedBranch(undefined);
    setShow(e);
  };

  const { errors, handleChange, resetValues, setValues, submit } = useForm({
    initialValues: {
      branch: '',
    },
    validationSchema: Yup.object({
      branch: Yup.string().required(),
    }),
    onSubmit: async () => {
      if (selectedBranch && show?.data?.repo) {
        onSubmit?.({
          branch: selectedBranch.value,
          repo: show?.data?.repo,
          provider: show.data.provider,
        });
        resetValues();
        onClose(null);
      }
    },
  });

  useEffect(() => {
    if (show) {
      switch (show.data?.provider) {
        case 'github':
          setGithubRepoUrl(show.data?.repo ? show.data.repo : null);
          break;
        case 'gitlab':
          setGitlabRepoUrl(show.data?.repo ? show.data.repo : null);
          break;
        default:
          break;
      }
      if (show.data?.branch) {
        setSelectedBranch({
          label: show?.data?.branch,
          value: show?.data?.branch,
        });
        setValues({ branch: show.data.branch });
      }
    }
  }, [show]);

  useEffect(() => {
    if (show?.data) {
      switch (show.data.provider) {
        case 'github':
          setFetchedBranches(
            githubBrances.data?.map((gb) => ({
              label: gb.name || '',
              value: gb.name || '',
            })) || []
          );
          break;
        case 'gitlab':
          setFetchedBranches(
            gitlabBrances.data?.map((gb) => ({
              label: gb.name || '',
              value: gb.name || '',
            })) || []
          );
          break;
        default:
          setFetchedBranches([]);
          break;
      }
    } else {
      setFetchedBranches([]);
    }
  }, [githubBrances.data, gitlabBrances.data]);

  return (
    <Pulsable isLoading={githubBrances.isLoading || gitlabBrances.isLoading}>
      <div className="mt-xl rounded border border-border-default shadow-button p-3xl flex flex-col gap-3xl">
        <Select
          label="Branch"
          placeholder="Select branch"
          options={async () => fetchedBranches}
          value={selectedBranch}
          onChange={(val) => {
            handleChange('branch')(dummyEvent(val.value));
            setSelectedBranch(val);
          }}
          message={errors.branch}
          error={!!errors.branch}
        />
        <div className="flex items-center justify-end gap-3xl">
          <Button
            content="Cancel"
            variant="basic"
            onClick={() => {
              onClose(null);
            }}
          />
          <Button
            content="Choose"
            variant="primary"
            onClick={() => {
              submit();
            }}
          />
        </div>
      </div>
    </Pulsable>
  );
};

// interface IMappedRepo {
//   name: string;
//   private: boolean;
//   updatedAt: string;
//   url: string;
// }
interface IListRenderer {
  response: SWRResponse<
    { name: string; updatedAt: any; private: true; url: string }[],
    any,
    any
  >;
  provider: 'github' | 'gitlab';
  // isLoading: boolean;
  // repos: IMappedRepo[];
  selectedBranch: IBranch | null | undefined;
  onShowBranch(data: IShowDialog<IBranch | null>): void;
  onImport(data: IBranch): void;
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

const ListRenderer = ({
  response,
  provider,
  selectedBranch,
  onShowBranch,
  onImport,
}: IListRenderer) => {
  const { isLoading, data } = response;

  return (
    <Pulsable isLoading={isLoading || !data}>
      <List.Root className="min-h-[356px]">
        {[...(isLoading || !data ? mockListData : data)].map((repo, index) => {
          return (
            <List.Row
              key={repo.name}
              columns={[
                {
                  key: generateKey(repo.name || '', index),
                  className: 'flex-1 pulsable',
                  render: () => (
                    <div className="flex flex-row gap-lg items-center bodyMd-medium flex-1">
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
                  ),
                },
                ...[
                  ...(selectedBranch && selectedBranch.repo === repo.url
                    ? [
                        {
                          key: generateKey(repo?.url || ''),
                          className: 'cursor-pointer',
                          label: (
                            <ButtonGroup.Root
                              selectable
                              value=""
                              onValueChange={() => {}}
                              onClick={() => {
                                onShowBranch?.({
                                  type: DIALOG_TYPE.NONE,
                                  data: {
                                    provider,
                                    repo: repo.url!,
                                    branch:
                                      selectedBranch?.repo === repo.url
                                        ? selectedBranch?.branch
                                        : null,
                                  },
                                });
                              }}
                            >
                              <ButtonGroup.IconButton
                                icon={
                                  <div className="flex gap-md items-center">
                                    <GitBranch size={16} />
                                    {selectedBranch.branch}
                                  </div>
                                }
                                value="1"
                              />
                              <ButtonGroup.IconButton
                                icon={<Pencil />}
                                value=""
                              />
                            </ButtonGroup.Root>
                          ),
                        },
                      ]
                    : []),
                ],
                {
                  key: generateKey(repo.name || '', 'action', index),
                  className: 'pulsable',
                  render: () => (
                    <Button
                      content={
                        selectedBranch?.repo === repo.url
                          ? 'import'
                          : 'Choose branch'
                      }
                      variant={
                        selectedBranch?.repo === repo.url ? 'primary' : 'basic'
                      }
                      onClick={() => {
                        if (selectedBranch?.repo !== repo.url) {
                          onShowBranch?.({
                            type: DIALOG_TYPE.NONE,
                            data: {
                              provider,
                              repo: repo.url!,
                              branch:
                                selectedBranch?.repo === repo.url
                                  ? selectedBranch?.branch
                                  : null,
                            },
                          });
                        } else {
                          onImport?.(selectedBranch);
                        }
                      }}
                    />
                  ),
                },
              ]}
            />
          );
        })}
      </List.Root>
    </Pulsable>
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
  onImport?(source: IBranch): void;
}
const GitRepoSelector = ({ onChange, onImport }: IGitRepoSelector) => {
  const githubInstallUrl = `https://github.com/apps/${githubAppName}/installations/new`;

  const [showBranchChooser, setShowBranchChooser] =
    useState<IShowDialog<IBranch | null>>(null);
  const [showProviderSwitch, setProviderSwitch] = useState(false);

  const [searchText, setSearchText] = useState('');

  const [provider, setProvider] = useState<'github' | 'gitlab'>('github');
  const {
    installations,
    repos: githubRepos,
    setOrganization,
    setSearchText: setGithubRepoSearchText,
    setFetchInstallations,
  } = useGithubProvider({});

  const {
    groups,
    repos: gitlabRepos,
    setFetchGroups,
    setGroupId,
    setSearchText: setGitlabRepoSearchText,
  } = useGitlabProvider({
    fetchGroups: false,
  });

  const [selectedBranch, setSelectedBranch] = useState<IBranch | null>();
  const [options, setOptions] = useState<IRepoRender[]>([]);

  const [selectedAccount, setSelectedAccount] = useState<IRepoRender>();

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
    // setSearchLoading(true);
  }, [searchText]);

  const valueRender = ({ label, labelValueIcon }: IRepoRender) => {
    return (
      <div className="flex flex-row gap-xl items-center bodyMd text-text-default">
        <span>{labelValueIcon}</span>
        <span>{label}</span>
      </div>
    );
  };

  // useEffect(() => {
  //   console.log(installations.data?.length || 0, installations.isLoading);
  // }, [installations]);

  return (
    <>
      {!showBranchChooser && (
        <div className="flex flex-col gap-6xl">
          <div className="headingXl text-text-default">
            Import Git Repository
          </div>
          <div className="flex flex-col gap-6xl relative">
            <div className="flex flex-row gap-lg items-center">
              <div className="flex-1">
                <Pulsable
                  isLoading={installations.isLoading || groups.isLoading}
                >
                  <div className="pulsable">
                    <Select
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
                <Pulsable
                  isLoading={installations.isLoading || groups.isLoading}
                >
                  <TextInput
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
              response={provider === 'github' ? githubRepos : gitlabRepos}
              provider={provider}
              selectedBranch={selectedBranch}
              onShowBranch={(data) => {
                setShowBranchChooser(data);
              }}
              onImport={(data) => {
                onImport?.(data);
              }}
            />
            <AnimatePresence mode="wait">
              {showProviderSwitch && (
                <motion.div
                  // initial={{ opacity: 0.85 }}
                  // animate={{ opacity: 1 }}
                  // exit={{ opacity: 0 }}
                  // transition={{ ease: 'linear' }}
                  className="absolute z-10 inset-0 flex flex-col items-center justify-center bg-surface-basic-subdued border border-border-default rounded"
                >
                  <div className="text-text-soft bodyMd mb-5xl">
                    Select a Git provider to import an existing project from a
                    Git Repository.
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
        </div>
      )}

      {showBranchChooser && (
        <BranchChooser
          show={showBranchChooser}
          setShow={setShowBranchChooser}
          onSubmit={(val) => {
            setSelectedBranch(val);
            onChange?.(val);
          }}
        />
      )}
    </>
  );
};

export default GitRepoSelector;
