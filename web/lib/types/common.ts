import { ReactNode } from 'react';

export type NonNullableString = string & NonNullable<unknown>;

export type MapType = {
  [key: string]: string | number | MapType;
};

export interface IChildren {
  children: ReactNode;
}

export interface IRemixHeader {
  get?: any;
}

export interface IRemixReq {
  headers: IRemixHeader;
  url: string;
  method: 'GET' | 'POST' | (string & NonNullable<unknown>);
  json: () => Promise<MapType>;
}

export interface IRemixCtx {
  request: IRemixReq;
  params: MapType;
}

interface IExtRemixReq extends IRemixReq {
  cookies: string[];
}
export interface IExtRemixCtx extends IRemixCtx {
  authProps: any;
  consoleContextProps: any;
  accounts: any;
  request: IExtRemixReq;
}

export type ICookie = any;
export type ICookies = ICookie[];

export interface IGQLServerHandler {
  headers: IRemixHeader;
  cookies?: ICookies;
}

export type IGqlReturn<T> = Promise<{
  errors?: Error[];
  data: T;
}>;
