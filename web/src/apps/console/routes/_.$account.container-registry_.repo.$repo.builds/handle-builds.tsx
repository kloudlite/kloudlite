import { ReactNode, useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { TextArea, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import GitRepoSelector from '~/console/components/git-repo-selector';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import { IDialog } from '~/console/components/types.d';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';

import {
    GitBranch,
    GithubLogoFill,
    GitlabLogoFill,
    MinusCircle,
    PencilSimple,
} from '@jengaicons/react';
import { useParams } from '@remix-run/react';
import { Checkbox } from '~/components/atoms/checkbox';
import Select from '~/components/atoms/select';
import { toast } from '~/components/molecule/toast';
import { uuid } from '~/components/utils';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { handleError } from '~/root/lib/utils/common';

interface IKeyValuePair {
  onChange?(item: Array<Record<string, any>>): void;
  value?: Array<Record<string, any>>;
  label?: ReactNode;
}
export const KeyValuePair = ({
  onChange,
  value = [],
  label,
}: IKeyValuePair) => {
  const newItem = [{ key: '', value: '', id: uuid() }];
  const [items, setItems] = useState<Array<Record<string, any>>>(newItem);

  const handleChange = (_value = '', id = '', target = {}) => {
    setItems(
      items.map((i) => {
        if (i.id === id) {
          switch (target) {
            case 'key':
              return { ...i, key: _value };
            case 'value':
            default:
              return { ...i, value: _value };
          }
        }
        return i;
      })
    );
  };

  useEffect(() => {
    if (onChange) onChange(Array.from(items));
  }, [items]);

  useEffect(() => {
    if (value.length > 0) {
      setItems(Array.from(value));
    }
  }, []);

  return (
    <div className="flex flex-col gap-xl">
      <div className="flex flex-col gap-md">
        {label && (
          <span className="text-text-default bodyMd-medium">{label}</span>
        )}
        {items.map((item) => (
          <div key={item.id} className="flex flex-row gap-xl items-end">
            <div className="flex-1">
              <TextInput
                placeholder="Key"
                value={item.key}
                onChange={({ target }) =>
                  handleChange(target.value, item.id, 'key')
                }
              />
            </div>
            <div className="flex-1">
              <TextInput
                placeholder="Value"
                value={item.value}
                onChange={({ target }) =>
                  handleChange(target.value, item.id, 'value')
                }
              />
            </div>
            <IconButton
              icon={<MinusCircle />}
              variant="plain"
              disabled={items.length < 2}
              onClick={() => {
                setItems(items.filter((i) => i.id !== item.id));
              }}
            />
          </div>
        ))}
      </div>
      <Button
        variant="basic"
        content="Add arg"
        size="sm"
        onClick={() => {
          setItems([...items, { ...newItem[0], id: uuid() }]);
        }}
      />
    </div>
  );
};

interface ISource {
  repo: string;
  branch: string;
  provider: 'github' | 'gitlab';
}

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

const HandleBuild = ({ show, setShow }: IDialog) => {
  const api = useConsoleApi();
  const reloadPage = useReload();

  const [source, setSource] = useState<ISource | null>(null);

  const [advanceOptions, setAdvanceOptions] = useState<
    boolean | undefined | string
  >(false);

  const { currentStep, onNext, onPrevious, reset } = useMultiStep({
    defaultStep: 1,
    totalSteps: 2,
  });

  const { repo } = useParams();

  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    resetValues,
    setValues,
    isLoading,
  } = useForm({
    initialValues: {
      name: '',
      tags: [],
      repository: repo,
      contextDir: '',
      dockerfilePath: '',
      dockerfileContent: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      tags: Yup.array()
        .required()
        .test('is-valid', 'Tags is required', (value) => {
          return value.length > 0;
        }),
    }),
    onSubmit: async (val) => {
      if (source) {
        try {
          const { errors: e } = await api.createBuild({
            build: {
              name: val.name,
              repository: val.repository || '',
              source: {
                branch: source.branch!,
                repository: source.repo!,
                provider: source.provider!,
              },
              tags: val.tags,
            },
          });
          if (e) {
            throw e[0];
          }
          reloadPage();
          resetValues();
          toast.success('Build created successfully');
          setShow(null);
        } catch (err) {
          handleError(err);
        }
      }
    },
  });

  const getProviderLogo = (provider: string) => {
    const logoSize = 24;
    switch (provider) {
      case 'github':
        return <GithubLogoFill size={logoSize} />;
      case 'gitlab':
        return <GitlabLogoFill size={logoSize} />;
      default:
        return null;
    }
  };

  return (
    <Popup.Root
      className="!w-[750px]"
      show={show as any}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>Create build</Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <MultiStep.Root currentStep={currentStep}>
            <MultiStep.Step step={1}>
              <div className="p-xl !pt-0">
                <GitRepoSelector
                  onImport={(val) => {
                    setSource({ ...val, branch: val.branch! });
                    onNext();
                  }}
                />
              </div>
            </MultiStep.Step>
            <MultiStep.Step step={2}>
              <div className="flex flex-col gap-2xl">
                <div className="flex flex-col gap-xl rounded border border-border-default p-xl">
                  <div className="flex flex-row gap-3xl items-center justify-between">
                    <div className="flex flex-row items-center gap-lg">
                      {getProviderLogo(source?.provider || '')}{' '}
                      <div className="bodyMd-medium">{source?.repo}</div>
                    </div>
                    <div className="flex flex-row items-center gap-lg">
                      <GitBranch size={16} />{' '}
                      <div className="bodyMd-medium">{source?.branch}</div>
                    </div>

                    <div className="self-end pt-lg">
                      <Button
                        content="Change"
                        variant="basic"
                        prefix={<PencilSimple />}
                        size="sm"
                        onClick={() => {
                          onPrevious();
                        }}
                      />
                    </div>
                  </div>
                </div>
                <div className="flex flex-col gap-xl rounded border border-border-default p-xl">
                  <div className="headingXl text-text-default">Target</div>
                  <div className="flex flex-col gap-xl">
                    <TextInput
                      label="Name"
                      value={values.name}
                      onChange={handleChange('name')}
                      error={!!errors.name}
                      message={errors.name}
                    />
                    <Select
                      label="Tags"
                      creatable
                      multiselect
                      options={[]}
                      value={values.tags.map((t) => ({ label: t, value: t }))}
                      onCreateOption={(val) => {
                        handleChange('tags')(dummyEvent([...values.tags, val]));
                      }}
                      error={!!errors.tags}
                      message={errors.tags}
                    />
                    <Checkbox
                      label="Advance options"
                      checked={advanceOptions}
                      onChange={setAdvanceOptions}
                    />
                    {advanceOptions && (
                      <div className="flex flex-col gap-xl">
                        <KeyValuePair label="Build args" value={[]} />
                        <KeyValuePair label="Build contexts" value={[]} />
                        <TextInput
                          label="Context dir"
                          value={values.contextDir}
                          onChange={handleChange('contextDir')}
                        />
                        <TextInput
                          label="Docker file path"
                          value={values.dockerfilePath}
                          onChange={handleChange('dockerfilePath')}
                        />
                        <TextArea
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
                </div>
              </div>
            </MultiStep.Step>
          </MultiStep.Root>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          {currentStep > 1 && (
            <Popup.Button
              type="submit"
              content={show?.type === 'add' ? 'Create' : 'Update'}
              variant="primary"
            />
          )}
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleBuild;
