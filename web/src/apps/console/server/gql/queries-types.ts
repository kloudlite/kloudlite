import { ISecret } from '../utils/kresources/secret';

type ParamsType = {
  [key: string]: ISecret;
};

type APIFunctionType = (
  params: ParamsType
) => Promise<{ data: ParamsType; errors: [Error] }>;

export type ApiType = {
  [key: string]: APIFunctionType;
};
