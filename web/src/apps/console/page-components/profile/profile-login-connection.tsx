import {
  ArrowRight,
  ArrowSquareOut,
  GithubLogoFill,
  GitlabLogoFill,
} from '@jengaicons/react';
import { useEffect } from 'react';
import { string } from 'yup';
import { Avatar } from '~/components/atoms/avatar';
import { IconButton } from '~/components/atoms/button';
import { generateKey } from '~/components/utils';
import { ListTitle } from '~/console/components/console-list-components';
import List from '~/console/components/list';
import ResourceExtraAction, {
  IResourceExtraItem,
} from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { isBrowser } from '~/root/lib/client/helpers/is-browser';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';

type ExtractKeyType<T> = keyof T;
const extractKeyType = <T,>({ key }: { object: T; key: string }) => {
  return key as ExtractKeyType<T>;
};

const LOGIN_CONNECTIONS = {
  providerGithub: {
    label: 'Github',
    icon: <GithubLogoFill size={24} />,
    avatar: string,
    url: 'github.com',
    loginUrl: '',
    urlKey: 'githubLoginUrl',
    // @ts-ignore
    manageUrl: isBrowser() ? window.MANAGE_GITHUB_URL : '',
  },
  providerGitlab: {
    label: 'Gitlab',
    icon: <GitlabLogoFill size={24} />,
    avatar: string,
    url: 'gitlab.com',
    loginUrl: '',
    urlKey: 'gitlabLoginUrl',
    // @ts-ignore
    manageUrl: isBrowser() ? window.MANAGE_GITLAB_URL : '',
  },
};

const ExtraButton = ({
  connection,
  isConnected,
}: {
  connection: (typeof LOGIN_CONNECTIONS)['providerGithub' | 'providerGitlab'];
  isConnected: boolean;
}) => {
  console.log('connection', connection);
  const items: IResourceExtraItem[] = [];

  if (isConnected) {
    items.push({
      label: `Manage on ${connection.url}`,
      suffix: <ArrowSquareOut size={16} />,
      type: 'item',
      to: connection.manageUrl,
      linkProps: {
        target: '_blank',
        rel: 'noopener noreferrer',
      },
      key: 'manage',
    });
    items.push({
      label: `Disconnect ${connection.label}`,
      type: 'item',
      key: 'disconnect',
      className: '!text-text-critical',
    });
  } else {
    items.push({
      label: `Connect ${connection.url}`,
      suffix: <ArrowSquareOut size={16} />,
      type: 'item',
      to: connection.loginUrl,
      linkProps: {
        target: '_blank',
        rel: 'noopener noreferrer',
      },
      key: 'manage',
    });
  }

  return <ResourceExtraAction options={items} />;
};

const ProfileLoginConnection = () => {
  const api = useConsoleApi();
  const { data: logins, error: errorLogins } = useCustomSwr(
    'api/login-connections/logins',
    async () => api.getLogins({})
  );
  const { data: loginUrls, error: errorLoginUrls } = useCustomSwr(
    'api/login-connections/loginUrls',
    async () => api.loginUrls({})
  );

  useEffect(() => {
    console.log('logins_conn', LOGIN_CONNECTIONS);
  }, []);

  return (
    <div className="flex flex-col gap-6xl">
      <div className="flex flex-col gap-md">
        <div className="headingMd text-text-default">Login connections</div>
        <div className="bodyMd text-text-soft">
          You can link your Personal Account on Kloudlite with a third-party
          service for login purposes. Please note that only one Login Connection
          can be added per third-party service.
        </div>
      </div>
      <List.Root
        plain
        className="rounded border border-border-default"
        header={
          <span className="bodyMd-medium text-text-default px-lg py-xl">
            Add new
          </span>
        }
      >
        {Object.entries(LOGIN_CONNECTIONS).map(([key, value]) => {
          const avatar =
            logins?.[extractKeyType({ object: logins, key })]?.avatar;
          const isConnected =
            !!logins?.[extractKeyType({ object: logins, key })];
          return (
            <List.Row
              key={generateKey(key)}
              className="!p-3xl min-h-[69px]"
              columns={[
                {
                  className: 'flex-1',
                  key: generateKey(key, value.label),
                  render: () => (
                    <div className="flex flex-row items-center gap-lg">
                      <ListTitle avatar={value.icon} />
                      <div className="flex flex-col items-center">
                        <div className="headingMd text-text-default">
                          {value.label}
                        </div>
                        <div>
                          {avatar && (
                            <Avatar
                              image={
                                <img
                                  src={avatar}
                                  alt={value.label}
                                  className="object-cover w-full h-full rounded-full"
                                />
                              }
                              size="xs"
                            />
                          )}
                        </div>
                      </div>
                    </div>
                  ),
                },
                {
                  key: generateKey(key, 'action'),
                  render: () =>
                    logins &&
                    (logins?.[extractKeyType({ object: logins, key })] ? (
                      <ExtraButton
                        connection={{
                          ...LOGIN_CONNECTIONS[
                            extractKeyType({ object: LOGIN_CONNECTIONS, key })
                          ],
                          loginUrl: isConnected
                            ? ''
                            : loginUrls?.[
                                extractKeyType({
                                  object: loginUrls,
                                  key: value.urlKey,
                                })
                              ],
                        }}
                        isConnected={isConnected}
                      />
                    ) : (
                      <IconButton
                        icon={<ArrowRight />}
                        variant="plain"
                        size="sm"
                      />
                    )),
                },
              ]}
            />
          );
        })}
      </List.Root>
    </div>
  );
};

export default ProfileLoginConnection;
