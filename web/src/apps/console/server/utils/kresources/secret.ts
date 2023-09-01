import { MetadataProps } from './common';

export type ISecretData = {
  [key: string]: string;
};

export interface ISecret {
  metadata: MetadataProps;
  displayName: string;
  stringData: ISecretData;
}

export const getMetadata = (
  { name, labels = {}, annotations = {}, namespace = undefined } = {
    name: '',
  }
) => ({
  ...{
    name,
    labels,
    annotations,
    namespace,
  },
});

export const getSecret = ({
  metadata,
  stringData,
  displayName,
}: ISecret): ISecret => ({
  ...{
    metadata,
    displayName,
    stringData,
  },
});
