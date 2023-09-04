import { Params } from '@remix-run/react';
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
