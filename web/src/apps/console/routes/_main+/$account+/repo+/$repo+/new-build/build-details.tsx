import { useEffect, useRef, useState } from 'react';
import { Checkbox } from '~/components/atoms/checkbox';
import { TextArea, TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import { useMapper } from '~/components/utils';
import KeyValuePair from '~/console/components/key-value-pair';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { dummyEvent } from '~/root/lib/client/hooks/use-form';

const BuildPlatforms = ({
  value,
  onChange,
}: {
  value?: Array<string>;
  onChange?(data: Array<string>): void;
}) => {
  const platforms = [
    { label: 'Arm', value: 'arm', checked: false },
    { label: 'x86', value: 'x86', checked: false },
    { label: 'x64', value: 'x64', checked: false },
  ];

  const [options, setOptions] = useState(platforms);

  useEffect(() => {
    setOptions((prev) =>
      prev.map((p) => {
        if (value?.includes(p.value)) {
          return { ...p, checked: true };
        }
        return { ...p, checked: false };
      })
    );
  }, [value]);

  useEffect(() => {
    onChange?.(options.filter((opt) => opt.checked).map((op) => op.value));
  }, [options]);

  return (
    <div className="flex flex-col gap-md">
      <span className="text-text-default bodyMd-medium">Platforms</span>
      <div className="flex flex-row items-center gap-xl">
        {options.map((bp) => {
          return (
            <Checkbox
              key={bp.label}
              label={bp.label}
              checked={bp.checked}
              onChange={(checked) => {
                setOptions((prev) =>
                  prev.map((p) => {
                    if (p.value === bp.value) {
                      return { ...p, checked: !!checked };
                    }
                    return p;
                  })
                );
              }}
            />
          );
        })}
      </div>
    </div>
  );
};

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

  useEffect(() => {
    console.log(values.tags);
  }, [values.tags]);

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
          console.log(val);
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
      <Checkbox
        label="Advance options"
        checked={values.advanceOptions}
        onChange={(check) => {
          handleChange('advanceOptions')(dummyEvent(!!check));
        }}
      />
      {values.advanceOptions && (
        <div className="flex flex-col gap-3xl">
          <KeyValuePair
            size="lg"
            label="Build args"
            value={Object.entries(values.buildArgs || {}).map(
              ([key, value]) => ({ key, value })
            )}
            onChange={(_, items) => {
              handleChange('buildArgs')(dummyEvent(items));
            }}
            error={!!errors.buildArgs}
            message={errors.buildArgs}
          />
          <KeyValuePair
            size="lg"
            label="Build contexts"
            value={Object.entries(values.buildContexts || {}).map(
              ([key, value]) => ({ key, value })
            )}
            onChange={(_, items) => {
              handleChange('buildContexts')(dummyEvent(items));
            }}
            error={!!errors.buildContexts}
            message={errors.buildContexts}
          />
          <TextInput
            size="lg"
            placeholder="Enter context dir"
            label="Context dir"
            value={values.contextDir}
            onChange={handleChange('contextDir')}
          />
          <TextInput
            size="lg"
            placeholder="Enter docker file path"
            label="Docker file path"
            value={values.dockerfilePath}
            onChange={handleChange('dockerfilePath')}
          />
          <TextArea
            placeholder="Enter docker file content"
            label="Docker file content"
            value={values.dockerfileContent}
            onChange={handleChange('dockerfileContent')}
            resize={false}
            rows="6"
          />
          <BuildPlatforms
            onChange={(data) => {
              console.log(data);
            }}
          />
        </div>
      )}
    </div>
  );
};

export default BuildDetails;
