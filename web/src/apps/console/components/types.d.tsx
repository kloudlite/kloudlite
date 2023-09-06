// Secret String Data
export type ISecretStringData = {
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
