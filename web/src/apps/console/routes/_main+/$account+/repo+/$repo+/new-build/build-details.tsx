import { useOutletContext } from '@remix-run/react';
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
import { constants } from '~/console/server/utils/constants';
// import KeyValuePairSelect from '~/console/components/key-value-pair-select';
import { IRepoContext } from '../_layout';

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

  const { repoName } = useOutletContext<IRepoContext>();
  console.log('reponame', repoName);

  const {
    data: digestData,
    isLoading: digestLoading,
    error: digestError,
  } = useCustomSwr(
    () =>
      constants.cacheRepoName ? `/digests_${constants.cacheRepoName}` : null,
    async () => {
      return api.listDigest({ repoName });
    }
  );

  const tags = useMapper(parseNodes(digestData), (val) => val.tags)
    .flat()
    .flatMap((f) => ({ label: f, value: f, updateInfo: null }));

  // const [isValidPath, setIsValidPath] = useState(false);

  console.log('tags', tags);

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
      {/* <div className="flex flex-row gap-3xl items-start">
        <div className="basis-1/3">
          <Select
            creatable
            size="lg"
            label="CacheKey"
            value={values.cacheKey}
            options={async () => tags}
            onChange={(_, val) => {
              handleChange('cacheKey')(dummyEvent(val));
            }}
            error={!!errors.digestData || !!digestError}
            message={
              errors.digestData || (digestError ? 'Error fetching tags.' : '')
            }
            loading={digestLoading}
            disableWhileLoading
          />
        </div>
        <div className="flex-1">
          <Select
            creatable
            size="lg"
            label="CachePath"
            value={values.cachePath}
            options={async () => []}
            onChange={(_, val) => {
              const pathRegex = /^\/(?:[\w.-]+\/)*(?:[\w.-]*)$/;
              if (pathRegex.test(val)) {
                setIsValidPath(true);
                handleChange('cachePath')(dummyEvent(val));
              }
            }}
            error={!isValidPath}
            message={!isValidPath ? 'Invalid path' : ''}
            disableWhileLoading
          />
        </div>
      </div> */}
      <Checkbox
        label="Advance options"
        checked={values.advanceOptions}
        onChange={(check) => {
          handleChange('advanceOptions')(dummyEvent(!!check));
        }}
      />
      {values.advanceOptions && (
        <div className="flex flex-col gap-3xl">
          {/* <KeyValuePairSelect
            size="lg"
            label="Caches"
            value={Object.entries(values.caches || {}).map(([key, value]) => ({
              key,
              value,
            }))}
            onChange={(_, items) => {
              handleChange('caches')(dummyEvent(items));
            }}
            error={!!errors.buildArgs}
            message={errors.buildArgs}
            options={tags}
          /> */}
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
