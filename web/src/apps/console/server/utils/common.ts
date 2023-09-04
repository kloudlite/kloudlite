import { Params } from '@remix-run/react';
import { decodeUrl } from '~/root/lib/client/hooks/use-search';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { IRemixCtx } from '~/root/lib/types/common';
import { ProjectId, WorkspaceOrEnvId } from '~/root/src/generated/gql/server';

interface IParamsCtx {
  params: Params<string>;
}

const getScopeQuery = (ctx: IParamsCtx): WorkspaceOrEnvId => {
  const { scope, workspace } = ctx.params;
  if (!workspace || !scope) {
    throw Error('scope and workspace is required, which is not provided');
  }
  return {
    value: workspace,
    type: scope === 'workspace' ? 'workspaceName' : 'environmentName',
  };
};

export const getProjectQuery = (ctx: IParamsCtx): ProjectId => {
  const { project } = ctx.params;
  if (!project) {
    throw Error(
      'project is required to render this page, which is not provide'
    );
  }
  return {
    type: 'name',
    value: project,
  };
};

export const getScopeAndProjectQuery = (
  ctx: IParamsCtx
): {
  project: ProjectId;
  scope: WorkspaceOrEnvId;
} => {
  return {
    project: getProjectQuery(ctx),
    scope: getScopeQuery(ctx),
  };
};

export const getPagination = (ctx: IRemixCtx) => {
  const { page } = getQueries(ctx);
  const { orderBy, sortDirection, last, first, before, after } =
    decodeUrl(page);

  return {
    ...{
      orderBy: orderBy || 'updateTime',
      sortDirection: sortDirection || 'DESC',
      last,
      first: first || (last ? undefined : 10),
      before,
      after,
    },
  };
};

export const getSearch = (ctx: IRemixCtx) => {
  const { search } = getQueries(ctx);
  const s = decodeUrl(search) || {};
  return {
    ...s,
  };
};

export const isValidRegex = (regexString = '') => {
  let isValid = true;
  try {
    RegExp(regexString);
  } catch (e) {
    isValid = false;
  }
  return isValid;
};
