import { ReactNode } from 'react';

export type MapType = {
  [key: string]: string | number | MapType;
};

export interface ChildrenProps {
  children: ReactNode;
}

export interface RHeaderProps {
  get?: any;
}

export interface RReqProps {
  headers: RHeaderProps;
  url: string;
  method: 'GET' | 'POST' | (string & NonNullable<unknown>);
  json: () => Promise<MapType>;
}

export interface RCtxProps {
  request: RReqProps;
  params: MapType;
}

interface ExtRReqProps extends RReqProps {
  cookies: string[];
}
export interface ExtRCtxProps extends RCtxProps {
  authProps: any;
  consoleContextProps: any;
  request: ExtRReqProps;
}

export type CookieType = any;
export type CookiesType = CookieType[];

export interface GQLServerHandlerProps {
  headers: RHeaderProps;
  cookies?: CookiesType;
}

export type GqlReturnProps<T> = Promise<{
  errors?: Error[];
  data: T;
}>;
