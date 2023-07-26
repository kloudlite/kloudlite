import { minimalAuth } from '~/root/lib/server/helpers/minimal-auth';
import { redirect } from 'react-router-dom';
import { getCookie } from '~/root/lib/app-setup/cookies';
import getQueries from '~/root/lib/server/helpers/get-queries';
import logger from '~/root/lib/client/helpers/log';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { GQLServerHandler } from '../gql/saved-queries';

// const restActions3 = async (ctx) => {
//   try {
//     const { data, errors } = await GQLServerHandler(ctx.request).listAccounts();
//     if (errors) {
//       throw errors[0];
//     }
//     if (data.length === 0) {
//       throw Error('no account created yet');
//     }
//     console.log(data);
//   } catch (err) {
//     return redirect('/new-account');
//   }
//
//   return false;
// };

const setProjectToContext = ({ ctx, project }) => {
  const cookie = getCookie(ctx);
  cookie.set('current_namespace', project.id);

  ctx.consoleContextProps = (props) => ({
    ...props,
    project,
    projects: project?.account?.projects,
    account: project.account,
  });
};

const setAccountToContext = ({ ctx, account }) => {
  const cookie = getCookie(ctx);

  cookie.set('kloudlite-account', account.id);

  ctx.consoleContextProps = (props) => ({
    ...props,
    account,
  });
};

export const doWithAccount = async ({ ctx, accountId }) => {
  if (!accountId) {
    return redirect(`${consoleBaseUrl}/projects`);
  }

  const { data, errors } = await GQLServerHandler(ctx.request).listProjects({});

  if (errors && errors.length) {
    logger.error('errors', errors);
  }

  if (!data) {
    return redirect(`${consoleBaseUrl}/projects`);
  }

  // logger.log('running', data, accountId);

  const { projects, account } = data || {};

  if (!projects || (projects && projects.length === 0)) {
    setAccountToContext({ ctx, account });

    // logger.log('no projects', account);
    // if (ctx.req.url === '/project/new') return false;

    // const { pathname } = new URL(ctx.request.url);
    // if (pathname === '/projects/create-project') return false;

    return redirect(`/new-project`, {
      headers: {
        'Content-Type': 'Application/json',
        'Set-Cookie': ctx.request.cookies || [],
      },
    });
  }

  return redirect(`${consoleBaseUrl}/?projectId=${projects[0].id}`, {
    headers: {
      'Content-Type': 'Application/json',
      'Set-Cookie': ctx.request.cookies || [],
    },
  });
  // return false;
};

const restActions = async (ctx) => {
  const { pathname } = new URL(ctx.request.url);
  // logger.log('pathname', pathname);
  if (pathname.startsWith('/projects') || pathname.startsWith('/clusters'))
    return false;

  const cookie = getCookie(ctx);
  const query = getQueries(ctx);
  // logger.log(query, ctx.request.url);

  const selectedAccountId = query.accountName;
  const currentProjectId = query.namespace || cookie.get('current_namespace');

  if (currentProjectId || selectedAccountId) {
    // logger.log('currentProjectId', currentProjectId, selectedAccountId);
    if (!currentProjectId) {
      logger.log('no namespace');

      const x = await doWithAccount({ ctx, accountId: selectedAccountId });
      return x;
    }

    const { data: projectData, errors } = await GQLServerHandler(
      ctx.request
    ).getProject({
      projectId: currentProjectId,
    });

    if (errors && errors.length) {
      logger.error('errors', errors);
    }

    if (
      (selectedAccountId && selectedAccountId !== projectData?.account?.id) ||
      !projectData?.id
    ) {
      return doWithAccount({ ctx, accountId: selectedAccountId });
    }

    if (errors && errors.length) {
      // setAccountToContext({ ctx, account });
      return redirect(`${consoleBaseUrl}/projects`);
    }
    setProjectToContext({ ctx, project: projectData });
    return false;
  }

  return redirect(`${consoleBaseUrl}/projects`);
};

export const setupConsoleContext = async (ctx) => {
  return (await minimalAuth(ctx)) || restActions(ctx);
};
