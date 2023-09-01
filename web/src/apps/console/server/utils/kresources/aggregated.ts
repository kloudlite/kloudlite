import { UserProps } from '../user';
import { ISecret } from './secret';
import { WorkspaceProps } from './workspace';

export type KubeResType = ISecret | WorkspaceProps;

export const parseName = (resource: KubeResType) =>
  resource?.metadata?.name || '';

export interface IClientContext {
  user?: UserProps;
  workspace?: WorkspaceProps;
}
