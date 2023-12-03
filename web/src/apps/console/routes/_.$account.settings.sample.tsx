import { useLoaderData } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import logger from '~/root/lib/client/helpers/log';
import { githubAppName } from '~/root/lib/configs/base-url.cjs';
import { IRemixCtx } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import GitRepoSelector from '../components/git-repo-selector';
import { IAccount } from '../server/gql/queries/access-queries';
import { ConsoleApiType, GQLServerHandler } from '../server/gql/saved-queries';
import { parseName } from '../server/r-utils/common';

export const popupWindow = ({
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
      // logger.log('closed');
      onClose();
    }
  }, 100);
};

export const loader = async (ctx: IRemixCtx) => {
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getLogins({});

    if (errors) {
      throw errors[0];
    }

    const { data: e, errors: dErrors } = await GQLServerHandler(
      ctx.request
    ).loginUrls({});

    if (dErrors) {
      throw dErrors[0];
    }

    return {
      loginUrls: e,
      logins: data,
    };
  } catch (err) {
    logger.error(err);
  }

  return {
    logins: {},
  };
};

const Logins = () => {
  const eventListner = (e: MessageEvent) => {
    logger.log(e?.data?.status);
    logger.log(e, 'e');
    if (e?.data?.status === 'success') {
      console.log(e.data, 'logged in');
      // setContainerData((s) => ({
      //   ...s,
      //   gitProvider: provider,
      // }));
    }
  };

  const { logins, loginUrls } = useLoaderData();
  console.log(logins);
  const githubInstallUrl = `https://github.com/apps/${githubAppName}/installations/new`;
  return (
    <div className="flex flex-col gap-md">
      <h2>Logins</h2>
      <Button
        content="Install New App"
        onClick={() =>
          popupWindow({
            url: githubInstallUrl,
          })
        }
      />
      <Button
        content={logins.providerGithub ? 'github' : ''}
        onClick={() => {
          popupWindow({
            url: loginUrls.githubLoginUrl,
            onClose: async () => {
              window.removeEventListener('message', eventListner);
            },
          });
        }}
      />
      <Button
        onClick={() => {
          popupWindow({
            url: loginUrls.gitlabLoginUrl,
            onClose: async () => {
              window.removeEventListener('message', eventListner);
            },
          });
        }}
        content={logins.providerGitlab ? 'gitlab' : ''}
      />
    </div>
  );
};

export const updateAccount = async ({
  api,
  data,
}: {
  api: ConsoleApiType;
  data: IAccount;
}) => {
  try {
    const { errors: e } = await api.updateAccount({
      account: {
        displayName: data.displayName,
        metadata: {
          name: parseName(data),
        },
        contactEmail: data.contactEmail,
        spec: data.spec,
      },
    });
    if (e) {
      throw e[0];
    }
  } catch (err) {
    handleError(err);
  }
};

const SettingGeneral = () => {
  return (
    <>
      <Logins />
      <GitRepoSelector />
    </>
  );
};

export default SettingGeneral;
