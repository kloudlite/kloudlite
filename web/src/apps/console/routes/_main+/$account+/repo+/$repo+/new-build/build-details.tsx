import { useEffect, useRef } from 'react';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { useMapper } from '~/components/utils';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { dummyEvent } from '~/root/lib/client/hooks/use-form';
import AdvancedOptions from '../builds/advanced-options';

const BuildDetails = ({
  values,
  handleChange,
  errors,
}: {
  values: { [key: string]: any };
  handleChange: (key: string) => (e: { target: { value: any } }) => void;
  errors: {
    [key: string]: string | undefined;
  };
}) => {
  const api = useConsoleApi();
  const { data: clusters, error: errorCluster } = useCustomSwr(
    '/clusters',
    async () => api.listClusters({})
  );

  const clusterData = useMapper(parseNodes(clusters), (item) => {
    return {
      label: item.displayName,
      value: parseName(item),
      render: () => (
        <div className="flex flex-col">
          <div>{item.displayName}</div>
          <div className="bodySm text-text-soft">{parseName(item)}</div>
        </div>
      ),
    };
  });

  const ref = useRef<HTMLInputElement>(null);
  useEffect(() => {
    ref.current?.focus();
  }, [ref.current]);

  return (
    <div className="flex flex-col gap-3xl">
      <TextInput
        ref={ref}
        label="Build name"
        size="lg"
        value={values.name}
        onChange={handleChange('name')}
        error={!!errors.name}
        message={errors.name}
      />
      <Select
        label="Tags"
        size="lg"
        placeholder="Add tags"
        creatable
        multiple
        value={values.tags}
        options={async () =>
          values.tags.map((t: string) => ({ label: t, value: t }))
        }
        onChange={(_, val) => {
          handleChange('tags')(dummyEvent(val));
        }}
        error={!!errors.tags}
        message={errors.tags}
      />
      <Select
        label="Clusters"
        placeholder="Choose a cluster"
        size="lg"
        value={values.buildClusterName}
        options={async () => clusterData}
        onChange={(_, val) => {
          handleChange('buildClusterName')(dummyEvent(val));
        }}
        error={!!errors.buildClusterName || !!errorCluster}
        message={
          errors.buildClusterName ||
          (errorCluster ? 'Error loading clusters' : '')
        }
      />
      <AdvancedOptions
        values={values}
        handleChange={handleChange}
        errors={errors}
      />
    </div>
  );
};

export default BuildDetails;
