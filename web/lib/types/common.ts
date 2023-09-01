import { ReactNode } from 'react';

export type NonNullableString = string & NonNullable<unknown>;

export type MapType<T = string | number | boolean> = {
  [key: string]: T | MapType<T>;
};

export type FlatMapType<T = string | number | boolean> = {
  [key: string]: T;
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
  cookies: string[];
}

export interface IRemixCtx {
  request: IRemixReq;
  params: FlatMapType<string>;
}

export interface IExtRemixCtx extends IRemixCtx {
  authProps: any;
  consoleContextProps: any;
  accounts: any;
  request: IRemixReq;
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
