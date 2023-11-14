import { IButton } from '~/components/atoms/button';

// Secret String Data
export type IConfigOrSecretData = {
  [key: string]: string;
};

// Modified Config or Secret value
export type ICSValueExtended = {
  value: string;
  insert: boolean;
  edit: boolean;
  delete: boolean;
  newvalue: string | null;
};

// Config or Secret Base structure
export type ICSBase = {
  key: string;
  value: ICSValueExtended;
};

// Modified Config or Secret Data Structure
export type IModifiedItem = {
  [key: string]: ICSValueExtended;
};

// Dialog state
export type IShowDialog<T = null> = {
  type: string;
  data: T;
} | null;

// dialog params
export interface IDialog<A = null, T = null> {
  show: IShowDialog<A>;
  setShow: React.Dispatch<React.SetStateAction<IShowDialog<A>>>;
  onSubmit?: (data: T) => void;
}

export interface ISubNavCallback {
  primaryAction?: IButton & { show: boolean };
  secondaryAction?: IButton & { show: boolean };
}
