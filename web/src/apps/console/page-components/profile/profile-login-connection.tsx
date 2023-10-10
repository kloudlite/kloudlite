import {
  ArrowRight,
  ArrowSquareOut,
  GithubLogoFill,
  GitlabLogoFill,
} from '@jengaicons/react';
import { IconButton } from '~/components/atoms/button';
import { generateKey } from '~/components/utils';
import { ListTitleWithAvatar } from '~/console/components/console-list-components';
import List from '~/console/components/list';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';

type ExtractKeyType<T> = keyof T;
const extractKeyType = <T,>({ key }: { object: T; key: string }) => {
  return key as ExtractKeyType<T>;
};

const LOGIN_CONNECTIONS = {
  providerGithub: {
    label: 'Github',
    icon: <GithubLogoFill size={24} />,
    manageUrl:
      'https://github.com/settings/connections/applications/Iv1.1a4fc72dd418358c',
    url: 'github.com',
  },
  providerGitlab: {
    label: 'Gitlab',
    icon: <GitlabLogoFill size={24} />,
    manageUrl: 'https://gitlab.com/-/profile/applications',
    url: 'gitlab.com',
  },
};

const ExtraButton = ({
  connection,
}: {
  connection: (typeof LOGIN_CONNECTIONS)['providerGithub' | 'providerGitlab'];
}) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: `Manage on ${connection.url}`,
          suffix: <ArrowSquareOut size={16} />,
          type: 'item',
          to: connection.manageUrl,
          linkProps: {
            target: '_blank',
            rel: 'noopener noreferrer',
          },
          key: 'manage',
        },
        {
          label: `Disconnect ${connection.label}`,
          type: 'item',
          key: 'disconnect',
          className: '!text-text-critical',
        },
      ]}
    />
  );
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
        {Object.entries(LOGIN_CONNECTIONS).map(([key, value]) => (
          <List.Row
            key={generateKey(key)}
            className="!p-3xl min-h-[69px]"
            columns={[
              {
                className: 'flex-1',
                key: generateKey(key, value.label),
                render: () => (
                  <ListTitleWithAvatar
                    title={value.label}
                    avatar={value.icon}
                  />
                ),
              },
              {
                key: generateKey(key, 'action'),
                render: () =>
                  logins &&
                  (logins?.[extractKeyType({ object: logins, key })] ? (
                    <ExtraButton
                      connection={
                        LOGIN_CONNECTIONS[
                          extractKeyType({ object: LOGIN_CONNECTIONS, key })
                        ]
                      }
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
        ))}
      </List.Root>
    </div>
  );
};

export default ProfileLoginConnection;
