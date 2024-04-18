import { useEffect, useState } from 'react';
import { Checkbox } from '~/components/atoms/checkbox';
import { TextArea, TextInput } from '~/components/atoms/input';
import { useMapper } from '~/components/utils';
import KeyValuePair from '~/iotconsole/components/key-value-pair';
import KeyValuePairSelect from '~/iotconsole/components/key-value-pair-select';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import { parseNodes } from '~/iotconsole/server/r-utils/common';
import { constants } from '~/iotconsole/server/utils/constants';
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

const AdvancedOptions = ({
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
  const api = useIotConsoleApi();
  const {
    data: digestData,
    isLoading: digestLoading,
    error: digestError,
  } = useCustomSwr(
    () =>
      constants.cacheRepoName ? `/digests_${constants.cacheRepoName}` : null,
    async () => {
      return api.listDigest({ repoName: constants.cacheRepoName });
    }
  );

  const tags = useMapper(parseNodes(digestData), (val) => val.tags)
    .flat()
    .flatMap((f) => ({ label: f, value: f, updateInfo: null }));

  return (
    <div className="flex flex-col gap-2xl">
      <Checkbox
        label="Advance options"
        checked={values.advanceOptions}
        onChange={(check) => {
          handleChange('advanceOptions')(dummyEvent(!!check));
        }}
      />
      {values.advanceOptions && (
        <div className="flex flex-col gap-3xl">
          <KeyValuePairSelect
            size="lg"
            label="Caches"
            value={values.caches || []}
            onChange={(val) => {
              handleChange('caches')(dummyEvent(val));
              console.log(val);
            }}
            error={!!errors.caches}
            message={errors.caches}
            selectError={!!errors.digestData || !!digestError}
            selectMessage={
              errors.digestData || (digestError ? 'Error fetching tags.' : '')
            }
            selectLoading={digestLoading}
            options={tags}
            keyLabel="name"
            valueLabel="path"
            regexPath={/^\/(?:[\w.-]+\/)*(?:[\w.-]*)$/}
          />
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

export default AdvancedOptions;
