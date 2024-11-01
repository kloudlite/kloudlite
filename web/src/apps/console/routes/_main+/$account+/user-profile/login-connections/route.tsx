import { Avatar } from '@kloudlite/design-system/atoms/avatar';
import { Button } from '@kloudlite/design-system/atoms/button';
import { generateKey } from '@kloudlite/design-system/utils';
import List from '~/console/components/list';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { popupWindow } from '~/console/utils/commons';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useExtLoaderData } from '~/root/lib/client/hooks/use-custom-loader-data';
import { gitEnvs } from '~/root/lib/configs/base-url.cjs';
import { IRemixCtx } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import ResourceExtraAction from '~/console/components/resource-extra-action';
import {
  ArrowsClockwise,
  GithubLogoFill,
  GitlabLogoFill,
  GearSix,
} from '~/console/components/icons';

type gitProvider = 'github' | 'gitlab';
export const loader = async (ctx: IRemixCtx) => {
  const { data, errors } = await GQLServerHandler(
    ctx.request
  ).getGitConnections({
    state: 'redirect:add-provider',
  });

  if (errors) {
    return handleError(errors[0]);
  }

  return {
    gitConnections: data,
  };
};

const LOGIN_CONNECTIONS = {
  github: {
    provider: 'github',
    label: 'Github',
    icon: <GithubLogoFill size={24} />,
    avatar: '',
    url: 'github.com',
    loginUrl: '',
    urlKey: 'githubLoginUrl',
    manageUrl: `https://github.com/apps/${gitEnvs.githubAppName}/installations/new`,
  },
  gitlab: {
    provider: 'gitlab',
    label: 'Gitlab',
    icon: <GitlabLogoFill size={24} />,
    avatar: '',
    url: 'gitlab.com',
    loginUrl: '',
    urlKey: 'gitlabLoginUrl',
    manageUrl: 'https://gitlab.com/-/profile/applications',
  },
};

type OnAction = ({ action }: { action: 'manage' | 'change' }) => void;

type IExtraButton = {
  onAction: OnAction;
};

const ExtraButton = ({ onAction }: IExtraButton) => {
  return (
    <ResourceExtraAction
      options={[
        {
          label: 'Manage',
          type: 'item',
          icon: <GearSix size={16} />,
          onClick: () => onAction({ action: 'manage' }),
          key: 'manage',
        },
        {
          label: 'Change',
          type: 'item',
          icon: <ArrowsClockwise size={16} />,
          onClick: () => onAction({ action: 'change' }),
          key: 'change',
        },
      ]}
    />
  );
};

const ProfileLoginConnection = () => {
  const { gitConnections } = useExtLoaderData<typeof loader>();
  const reloadPage = useReload();

  const addGitConnection = (provider: gitProvider) => {
    // window.addEventListener('message', eventListner);
    switch (provider) {
      case 'github':
        popupWindow({
          url: gitConnections.github.githubLoginUrl,
          onClose: () => {
            reloadPage();
          },
        });

        break;
      case 'gitlab':
        popupWindow({
          url: gitConnections?.gitlab.gitlabLoginUrl,
          onClose: () => {
            reloadPage();
          },
        });
        break;
      default:
        break;
    }
  };

  const manageGitConnection = (provider: gitProvider) => {
    switch (provider) {
      case 'github':
        popupWindow({
          url: LOGIN_CONNECTIONS.github.manageUrl,
          onClose: () => {
            reloadPage();
          },
        });

        break;
      case 'gitlab':
        popupWindow({
          url: LOGIN_CONNECTIONS.gitlab.manageUrl,
          onClose: () => {
            reloadPage();
          },
        });
        break;
      default:
        break;
    }
  };

  const onAction =
    (pv: gitProvider) =>
    ({ action }: { action: 'manage' | 'change' }) => {
      switch (action) {
        case 'manage':
          manageGitConnection(pv);
          break;
        case 'change':
          addGitConnection(pv);
          break;
        default:
          break;
      }
    };

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
          const data =
            value.provider === 'github'
              ? gitConnections?.github
              : gitConnections?.gitlab;

          const { provider } = value as {
            provider: gitProvider;
          };

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
                      <div className="flex items-center">
                        <span className="relative left-xs">{value.icon}</span>
                        <div className="h-sm w-xl bg-surface-primary-default" />
                        {data.avatar && (
                          <div className="border rounded-full">
                            <Avatar
                              image={
                                <img
                                  src={data.avatar}
                                  alt={value.label}
                                  className="object-cover w-full h-full rounded-full"
                                />
                              }
                              size="xs"
                            />
                          </div>
                        )}
                      </div>

                      <div className="headingMd text-text-default">
                        {value.label}
                      </div>
                    </div>
                  ),
                },
                {
                  key: generateKey(key, 'action'),
                  render: () => {
                    return data?.connected ? (
                      <ExtraButton onAction={onAction(provider)} />
                    ) : (
                      <Button
                        variant="primary-plain"
                        content="Connect"
                        onClick={() => addGitConnection(provider)}
                      />
                    );
                  },
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
