import { UserProps } from './user';
import { WorkspaceProps } from './workspace';

export type MapType = {
  [key: string]: string | number | MapType;
};

export interface MetadataProps {
  name: string;
  namespace?: string;
  labels: MapType;
  annotations: MapType;
}

export interface ContextProps {
  user?: UserProps;
  workspace?: WorkspaceProps;
}
