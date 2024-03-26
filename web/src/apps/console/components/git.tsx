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
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { dayjs } from '~/components/molecule/dayjs';
import Radio from '~/components/atoms/radio';
import { useAppend, useMapper } from '~/components/utils';
import { ReactNode, useEffect, useState } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { Button } from '~/components/atoms/button';
import { toast } from '~/components/molecule/toast';
import { gitEnvs } from '~/root/lib/configs/base-url.cjs';
import { ILoginUrls, ILogins } from '../server/gql/queries/git-queries';
import Pulsable from './pulsable';
import useGit, { IGIT_PROVIDERS } from '../hooks/use-git';

const extraAddOption = 'add';
const extraSwitchOption = 'switch';

const iconSize = 16;

const gitHubOptions = [
  {
    label: 'Add Github Account',
    value: extraAddOption,
    labelValueIcon: <Plus size={iconSize} />,
    render: () => (
      <div className="flex flex-row gap-lg items-center">
        <div>
          <Plus size={iconSize} />
        </div>
        <div>Add Github Account</div>
      </div>
    ),
  },
];

const commonOptions = [
  {
    label: 'Switch Git Provider',
    value: extraSwitchOption,
    labelValueIcon: <ListBullets size={iconSize} />,
    render: () => (
      <div className="flex flex-row gap-lg items-center">
        <div>
          <ListBullets size={iconSize} />
        </div>
        <div>Switch Git Provider</div>
      </div>
    ),
  },
];

const popupWindow = ({
  url = '',
  onClose = () => {},
  width = 800,
  height = 500,
  title = 'kloudlite',
}) => {
  const frame = window.open(
    url,
    title,
    `toolbar=no,scrollbars=yes,resizable=no,top=${
      window.screen.height / 2 - height / 2
    },left=${window.screen.width / 2 - width / 2},width=800,height=600`
  );
  const interval = setInterval(() => {
    if (frame && frame.closed) {
      clearInterval(interval);
      onClose();
    }
  }, 100);
};

interface IBranch {
  repo: string;
  branch: string;
  provider: IGIT_PROVIDERS;
}

interface IListRenderer {
  data:
    | { name: string; updatedAt: any; private: true; url: string }[]
    | undefined;
  onChange: (value: string) => void;
  value: string;
  isLoading?: boolean;
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

const ListRenderer = ({ data, onChange, value, isLoading }: IListRenderer) => {
  return (
    <div className="flex flex-col min-h-[260px]">
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
                className="flex-row justify-between w-full py-2xl bodyMd-medium"
                labelPlacement="left"
              >
                <div className="flex flex-row items-center gap-lg bodyMd-medium">
                  <span className="pulsable">{repo.name}</span>

                  <span className="pulsable">
                    {repo.private ? (
                      <LockSimple size={12} />
                    ) : (
                      <LockSimpleOpen size={12} />
                    )}
                  </span>
                  <span>
                    <CircleFill size={2} />
                  </span>
                  <span className="text-text-soft pulsable">
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
  label: string | undefined;
  labelValueIcon: JSX.Element;
  value: string;
  render: () => JSX.Element;
}

interface IGitRepoSelector {
  onChange?(source: IBranch): void;
  error?: string;
  value?: IBranch;
  logins: ILogins;
  loginUrls: ILoginUrls;
}

const valueRender = ({ label, labelValueIcon }: IRepoRender): ReactNode => {
  return (
    <div className="flex flex-row gap-xl items-center bodyMd text-text-default">
      <span>{labelValueIcon}</span>
      <span>{label}</span>
    </div>
  );
};

const Git = ({
  onChange,
  error,
  value,
  logins,
  loginUrls,
}: IGitRepoSelector) => {
  const [showProviderOverlay, setShowProviderOverlay] = useState(
    !logins?.providerGithub && !logins?.providerGitlab
  );

  const {
    searchText,
    setSearchText,
    installations,
    repos,
    repo,
    setRepo,
    branches,
    branch,
    setBranch,
    org,
    setOrg,
    provider,
    isLoading,
    forceFetchInstallations,
    revalidate,
  } = useGit({
    provider: value?.provider || 'github',
    retryWithOtherGitProvider: true,
    defaultFetch: true,
    onChange: (e) => {
      if (!showProviderOverlay) {
        onChange?.(e);
      } else {
        onChange?.({ repo: '', branch: '', provider });
      }
    },
  });

  const [loading, setLoading] = useState(true);

  const [githubButtonLoading, setGithubButtonLoading] = useState(false);
  const [gitlabButtonLoading, setGitlabButtonLoading] = useState(false);

  const accounts = useMapper(installations.data || [], (d) => {
    return {
      label: d.label,
      value: d.value,
      labelValueIcon:
        provider === 'gitlab' ? (
          <GitlabLogoFill size={iconSize} />
        ) : (
          <GithubLogoFill size={iconSize} />
        ),
      render: () => (
        <div className="flex flex-row gap-lg items-center">
          <div>
            {provider === 'gitlab' ? (
              <GitlabLogoFill size={iconSize} />
            ) : (
              <GithubLogoFill size={iconSize} />
            )}
          </div>
          <div>{d.label}</div>
        </div>
      ),
    };
  });

  const accountsModified = useAppend(
    accounts,
    provider === 'github'
      ? [...gitHubOptions, ...commonOptions]
      : [...commonOptions]
  );

  useEffect(() => {
    setShowProviderOverlay(!!installations.error && !isLoading);
  }, [installations.error, isLoading, installations.data]);

  useEffect(() => {
    setGitlabButtonLoading(false);
    setGithubButtonLoading(false);
  }, [installations.data, isLoading]);

  useEffect(() => {
    window.addEventListener('message', (e) => {
      if (e.data) {
        const { type, status, provider } = e.data;
        if (type === 'add-provider' && status === 'success') {
          revalidate(provider);
        }

        if (type === 'install') {
          revalidate(provider);
        }
      }
    });
    setTimeout(() => {
      setLoading(false);
    }, 200);
  }, []);

  useEffect(() => {
    if (showProviderOverlay) {
      setBranch('');
    }
  }, [showProviderOverlay]);

  useEffect(() => {
    if (!!installations.error || !!repos.error || !!branches.error) {
      toast.error(installations.error || repos.error || branches.error);
    }
  }, [installations.error, repos.error, branches.error]);

  return (
    <div>
      <div className="flex flex-col relative border border-border-default rounded px-2xl">
        <div className="flex flex-row gap-lg items-center py-2xl">
          <div className="flex-1">
            <Pulsable isLoading={installations.isLoading || loading}>
              <div className="pulsable">
                <Select
                  size="lg"
                  valueRender={valueRender}
                  options={async () => accountsModified}
                  value={org?.value}
                  onChange={(res) => {
                    switch (res.value) {
                      case extraAddOption:
                        popupWindow({
                          url: `https://github.com/apps/${gitEnvs.githubAppName}/installations/new`,
                        });
                        break;
                      case extraSwitchOption:
                        setShowProviderOverlay(true);
                        break;
                      default:
                        setOrg(res);
                        break;
                    }
                  }}
                />
              </div>
            </Pulsable>
          </div>
          <div className="flex-1">
            <Pulsable isLoading={installations.isLoading || loading}>
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
          data={repos.data}
          isLoading={installations.isLoading || repos.isLoading || loading}
          onChange={(v) => {
            setRepo(v);
          }}
        />

        <AnimatePresence mode="wait">
          {showProviderOverlay && (
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
                  loading={githubButtonLoading}
                  onClick={async () => {
                    setGithubButtonLoading(true);
                    const { errors } = await forceFetchInstallations('github');
                    if (errors) {
                      popupWindow({
                        url: loginUrls.githubLoginUrl,
                      });
                    } else if (provider === 'github') {
                      setShowProviderOverlay(false);
                      setGitlabButtonLoading(false);
                      setGithubButtonLoading(false);
                    } else {
                      revalidate('github');
                    }
                  }}
                />
                <Button
                  variant="purple"
                  content="Continue with Gitlab"
                  prefix={<GitlabLogoFill />}
                  block
                  loading={gitlabButtonLoading}
                  onClick={async () => {
                    setGitlabButtonLoading(true);
                    const { errors } = await forceFetchInstallations('gitlab');
                    if (errors) {
                      popupWindow({
                        url: loginUrls.gitlabLoginUrl,
                      });
                    } else if (provider === 'gitlab') {
                      setShowProviderOverlay(false);
                      setGitlabButtonLoading(false);
                      setGithubButtonLoading(false);
                    } else {
                      revalidate('gitlab');
                    }
                  }}
                />
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
      <div className="pt-3xl">
        <Pulsable
          isLoading={
            installations.isLoading ||
            repos.isLoading ||
            branches.isLoading ||
            branches.error ||
            repos.error ||
            installations.error ||
            loading
          }
        >
          <Select
            label="Select branch"
            size="lg"
            value={branch && !showProviderOverlay ? branch : undefined}
            disabled={!repo || showProviderOverlay}
            placeholder="Select a branch"
            options={async () => [
              ...(branches.data?.map((d) => ({
                label: d.name || '',
                value: d.name || '',
              })) || []),
            ]}
            onChange={({ value }) => {
              setBranch(value);
            }}
            error={!!error}
            message={error}
          />
        </Pulsable>
      </div>
    </div>
  );
};

export default Git;
