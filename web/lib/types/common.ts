import { ReactNode } from 'react';

export type MapType = {
  [key: string]: string | number | MapType;
};

export interface IChildren {
  children: ReactNode;
}

export interface IRHeader {
  get?: any;
}

export interface IRReq {
  headers: IRHeader;
  url: string;
  method: 'GET' | 'POST' | (string & NonNullable<unknown>);
  json: () => Promise<MapType>;
}

export interface IRCtx {
  request: IRReq;
  params: MapType;
}

interface IExtRReq extends IRReq {
  cookies: string[];
}
export interface IExtRCtx extends IRCtx {
  authProps: any;
  consoleContextProps: any;
  request: IExtRReq;
}

export type ICookie = any;
export type ICookies = ICookie[];

export interface IGQLServerHandler {
  headers: IRHeader;
  cookies?: ICookies;
}

export type IGqlReturn<T> = Promise<{
  errors?: Error[];
  data: T;
}>;
