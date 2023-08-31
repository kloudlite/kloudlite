import { MapType } from '~/root/lib/types/common';
import { UserProps } from './user';
import { WorkspaceProps } from './workspace';

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
