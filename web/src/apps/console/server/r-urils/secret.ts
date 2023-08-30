import { MetadataProps } from '../types/common';

export type SecretDataType = {
  [key: string]: string;
};

export interface SecretProps {
  metadata: MetadataProps;
  displayName: string;
  stringData: SecretDataType;
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
}: SecretProps): SecretProps => ({
  ...{
    metadata,
    displayName,
    stringData,
  },
});
