export const getMetadata = (
  { name, labels = [], annotations = [], namespace } = {
    name: '',
    namespace: '',
  }
) => ({
  ...{
    name,
    labels,
    annotations,
    namespace,
  },
});
