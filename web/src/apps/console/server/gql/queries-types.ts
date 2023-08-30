import { SecretProps } from '../r-urils/secret';

type ParamsType = {
  [key: string]: SecretProps;
};

type APIFunctionType = (
  params: ParamsType
) => Promise<{ data: ParamsType; errors: [Error] }>;

export type ApiType = {
  [key: string]: APIFunctionType;
};
