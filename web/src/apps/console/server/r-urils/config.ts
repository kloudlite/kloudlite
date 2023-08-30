type Data = {
  [key: string]: string;
};

interface Config {
  metadata: unknown;
  displayName: string;
  data: Data;
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

export const getConfig = ({ metadata, data, displayName }: Config): Config => ({
  ...{
    metadata,
    displayName,
    data,
  },
});
