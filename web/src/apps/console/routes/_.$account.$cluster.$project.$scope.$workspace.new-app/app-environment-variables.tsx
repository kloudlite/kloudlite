import {
  ArrowLeft,
  ArrowRight,
  LockSimple,
  LockSimpleOpen,
  X,
  XCircleFill,
} from '@jengaicons/react';
import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { Chip, ChipGroup } from '~/components/atoms/chips';
import { TextInput } from '~/components/atoms/input';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { useParams } from '@remix-run/react';
import { AnimatePresence, motion } from 'framer-motion';
import { handleError } from '~/root/lib/utils/common';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { NonNullableString } from '~/root/lib/types/common';
import List from '~/console/components/list';
import SelectPrimitive from '~/components/atoms/select-primitive';
import { IShowDialog } from '~/console/components/types.d';
import AppDialog from './app-dialogs';
import { FadeIn } from './util';
import { createAppEnvPage, useAppState } from './states';

interface IEnvVariable {
  key: string;
  type?: 'config' | 'secret' | NonNullableString;
  value?: string;
  refName?: string;
  refKey?: string;
}

interface IEnvVariablesList {
  envVariables: Array<IEnvVariable>;
  onDelete: (envVariable: IEnvVariable) => void;
}

export interface IValue {
  refKey: string;
  refName: string;
  type: 'config' | 'secret' | NonNullableString;
}

const EnvironmentVariablesList = ({
  envVariables,
  onDelete = (_) => _,
}: IEnvVariablesList) => {
  return (
    <div className="flex flex-col gap-lg">
      <div className="text-text-strong bodyMd">Environment variable list</div>
      <List.Root>
        {envVariables.map((ev, index) => {
          return (
            <List.Row
              key={ev.key}
              columns={[
                {
                  key: `${index}-column-0`,
                  render: () => (
                    <div className="text-icon-default">
                      {ev.type === 'config' && (
                        <LockSimpleOpen color="currentColor" size={16} />
                      )}
                      {ev.type === 'secret' && (
                        <LockSimple color="currentColor" size={16} />
                      )}
                    </div>
                  ),
                },
                {
                  key: `${index}-column-1`,
                  className: 'flex-1',
                  render: () => (
                    <div className="bodyMd-semibold text-text-default">
                      {ev.key}
                    </div>
                  ),
                },
                {
                  key: `${index}-column-2`,
                  className: 'flex-1',
                  render: () => (
                    <div className="flex flex-row gap-md items-center bodyMd text-text-soft">
                      {ev.type === undefined && ev.value}
                      {ev.type !== undefined && (
                        <>
                          {ev.refName}
                          <ArrowRight size={16} weight={1} />
                          {ev.refKey}
                        </>
                      )}
                    </div>
                  ),
                },
                {
                  key: `${index}-column-3`,
                  render: () => (
                    <div>
                      <IconButton
                        icon={<X />}
                        variant="plain"
                        size="sm"
                        onClick={() => {
                          onDelete(ev);
                        }}
                      />
                    </div>
                  ),
                },
              ]}
            />
          );
        })}
      </List.Root>
    </div>
  );
};

export const EnvironmentVariables = () => {
  const { setContainer, getContainer } = useAppState();

  const [showCSDialog, setShowCSDialog] = useState<IShowDialog>(null);
  // const [textInputValue, setTextInputValue] = useState<string>('');
  // const [value, setValue] = useState<IValue | null>(null);
  // const [key, setKey] = useState<string>('');
  // const [keyValueError, setKeyValueError] = useState<string | null>(null);

  const entry = Yup.object({
    type: Yup.string().oneOf(['config', 'secret']),
    key: Yup.string().required(),

    value: Yup.string().when(['type'], ([type], schema) => {
      if (type === undefined) {
        return schema.required();
      }
      return schema;
    }),
    refKey: Yup.string().when(['type'], ([type], schema) => {
      if (type === 'config' || type === 'secret') {
        return schema.required();
      }
      return schema;
    }),
    refName: Yup.string().when(['type'], ([type], schema) => {
      if (type === 'config' || type === 'secret') {
        return schema.required();
      }
      return schema;
    }),
  });

  const { values, setValues, submit } = useForm({
    initialValues: getContainer().env || [],
    validationSchema: Yup.array(entry),
    onSubmit: (val) => {
      setContainer((c) => ({
        ...c,
        env: val,
      }));
    },
  });
  useEffect(() => {
    submit();
  }, [values]);

  const addEntry = (val: IEnvVariable) => {
    setValues((v) => {
      v?.push({
        key: val.key,
        type: val.type,
        refName: val.refName || '',
        refKey: val.refKey || '',
        value: val.value || '',
      });
      return v;
    });
  };

  const removeEntry = (val: IEnvVariable) => {
    setValues((v) => {
      const nv = v?.filter((v) => v.key !== val.key);
      return nv;
    });
  };

  const vSchema = Yup.object({
    refKey: Yup.string().when('$textInputValue', ([v], schema) => {
      console.log(v, 'here');
      return schema;
    }),
    refName: Yup.string().when('$textInputValue', ([v], schema) => {
      console.log(v);
      return schema;
    }),
  });

  interface InitialValuesProps {
    key: string;
    value?: IValue;
    textInputValue?: string;
  }

  const initialValues: InitialValuesProps = {
    key: '',
    textInputValue: '',
  };

  const {
    values: eValues,
    errors: eErrors,
    handleChange: eHandleChange,
    setValues: eSetValues,
    resetValues,
    submit: eSubmit,
  } = useForm({
    initialValues,

    validationSchema: Yup.object().shape({
      value: vSchema,
      textInputValue: Yup.string(),
      key: Yup.string()
        .required()
        .test('is-valid', 'Key already exists.', (value) => {
          return !getContainer().env?.find((v) => v.key === value);
        }),
    }),
    onSubmit: () => {
      if (eValues.textInputValue) {
        const ev: IEnvVariable = {
          key: eValues.key,
          refKey: undefined,
          refName: undefined,
          type: undefined,
          value: eValues.textInputValue,
        };
        // setEnvVariables((prev) => [...prev, ev]);
        addEntry(ev);
      } else if (eValues.value) {
        const ev: IEnvVariable = {
          key: eValues.key,
          refKey: eValues.value.refKey,
          refName: eValues.value.refName,
          type: eValues.value.type,
          value: undefined,
        };

        // setEnvVariables((prev) => [...prev, ev]);
        addEntry(ev);
      }
      resetValues();
    },
  });

  return (
    <>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-default">
        <div className="flex flex-row gap-3xl items-start">
          <div className="flex-1">
            <TextInput
              label="Key"
              size="lg"
              error={!!eErrors.key}
              message={eErrors.key}
              value={eValues.key}
              autoComplete="off"
              onChange={eHandleChange('key')}
            />
          </div>
          <div className="flex-1">
            {eValues.value ? (
              <div className="flex flex-col gap-md">
                <div className="bodyMd-medium text-text-default">Value</div>
                <div className="flex flex-row items-center rounded border border-border-default bg-surface-basic-defaul">
                  <span className="py-lg pl-lg pr-md text-text-default">
                    {eValues.value.type === 'config' ? (
                      <LockSimpleOpen
                        size={14}
                        weight={2.5}
                        color="currentColor"
                      />
                    ) : (
                      <LockSimple size={14} weight={2.5} color="currentColor" />
                    )}
                  </span>
                  <div className="flex-1 flex flex-row gap-md items-center py-xl px-lg bodyMd text-text-soft">
                    {eValues.value.refKey}
                    <ArrowRight size={16} weight={1} />
                    {eValues.value.refName}
                  </div>
                  <button
                    tabIndex={-1}
                    type="button"
                    className="outline-none p-lg text-text-default rounded-full"
                    onClick={() => {
                      eSetValues((v) => {
                        return {
                          ...v,
                          value: undefined,
                        };
                      });
                    }}
                  >
                    <XCircleFill size={16} color="currentColor" />
                  </button>
                </div>
              </div>
            ) : (
              <TextInput
                value={eValues.textInputValue}
                onChange={eHandleChange('textInputValue')}
                label="Value"
                size="lg"
                suffix={
                  !eValues.textInputValue ? (
                    <ChipGroup
                      onClick={(data) => {
                        setShowCSDialog({ type: data.name, data: null });
                      }}
                    >
                      <Chip
                        label="Config"
                        item={{ name: 'config' }}
                        type="CLICKABLE"
                      />
                      <Chip
                        item={{ name: 'secret' }}
                        label="Secrets"
                        type="CLICKABLE"
                      />
                    </ChipGroup>
                  ) : null
                }
                showclear={!!eValues.textInputValue}
              />
            )}
          </div>
        </div>
        <div className="flex flex-row gap-md items-center">
          <div className="bodySm text-text-soft flex-1">
            All environment entries be mounted on the path specified in the
            container
          </div>
          <Button
            content="Add environment"
            variant="basic"
            disabled={
              !eValues.key || !(eValues.value || eValues.textInputValue)
            }
            onClick={() => {
              eSubmit();
            }}
          />
        </div>
      </div>
      {!!getContainer().env?.length && (
        <EnvironmentVariablesList
          envVariables={getContainer().env || []}
          onDelete={(ev) => {
            removeEntry(ev);
            // setEnvVariables((prev) => prev.filter((p) => p !== ev));
          }}
        />
      )}
      <AppDialog
        show={showCSDialog}
        setShow={setShowCSDialog}
        onSubmit={(item) => {
          eSetValues((v) => {
            return {
              ...v,
              value: item,
            };
          });
          setShowCSDialog(null);
        }}
      />
    </>
  );
};

interface IConfigMount {
  mountPath: string;
  refName: string;
}
interface IConfigMountList {
  configMounts: Array<IConfigMount>;
  onDelete: (configMount: IConfigMount) => void;
}
const ConfigMountsList = ({ configMounts, onDelete }: IConfigMountList) => {
  return (
    <div className="flex flex-col gap-lg">
      <div className="text-text-strong bodyMd">Config mount list</div>
      <List.Root>
        {configMounts.map((cm, index) => {
          return (
            <List.Row
              key={`${cm.mountPath} ${cm.refName}`}
              columns={[
                {
                  key: `${index}-column-0`,
                  render: () => (
                    <div className="text-icon-default">
                      <LockSimpleOpen color="currentColor" size={16} />
                    </div>
                  ),
                },
                {
                  key: `${index}-column-1`,
                  className: 'flex-1',
                  render: () => (
                    <div className="bodyMd-semibold text-text-default">
                      {cm.mountPath}
                    </div>
                  ),
                },
                {
                  key: `${index}-column-2`,
                  className: 'flex-1',
                  render: () => (
                    <div className="flex flex-row gap-md items-center bodyMd text-text-soft">
                      {cm.refName}
                    </div>
                  ),
                },
                {
                  key: `${index}-column-3`,
                  render: () => (
                    <div>
                      <IconButton
                        icon={<X />}
                        variant="plain"
                        size="sm"
                        onClick={() => {
                          onDelete(cm);
                        }}
                      />
                    </div>
                  ),
                },
              ]}
            />
          );
        })}
      </List.Root>
    </div>
  );
};

interface IConfigMounts {
  configMounts: Array<IConfigMount>;
  setConfigMounts: React.Dispatch<React.SetStateAction<Array<IConfigMount>>>;
}

const ConfigMounts = ({ configMounts, setConfigMounts }: IConfigMounts) => {
  const api = useAPIClient();

  const [isloading, setIsloading] = useState<boolean>(true);
  const { workspace, project, scope } = useParams();
  const [configs, setConfigs] = useState([]);

  const [mountPath, setMountPath] = useState<string>('');
  const [refName, setRefName] = useState<string>('');

  useDebounce(
    async () => {
      try {
        setIsloading(true);
        const { data, errors } = await api.listConfigs({
          project: {
            value: project,
            type: 'name',
          },
          scope: {
            value: workspace,
            type: scope === 'workspace' ? 'workspaceName' : 'environmentName',
          },
          //   pagination: getPagination(ctx),
          //   search: getSearch(ctx),
        });
        if (errors) {
          throw errors[0];
        }
        setConfigs(parseNodes(data));
      } catch (err) {
        handleError(err);
      } finally {
        setIsloading(false);
      }
    },
    300,
    []
  );
  return (
    <>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-default">
        <div className="flex flex-row gap-3xl items-center">
          <div className="flex-1">
            <TextInput
              label="Path"
              size="lg"
              autoComplete="off"
              value={mountPath}
              onChange={({ target }) => {
                setMountPath(target.value);
              }}
            />
          </div>
          <div className="flex-1">
            <SelectPrimitive.Root
              disabled={isloading}
              label="Config"
              value={refName}
              onChange={({ target }: any) => {
                setRefName(target.value);
              }}
            >
              <SelectPrimitive.Option value="">
                --- Select config ---
              </SelectPrimitive.Option>
              {configs.map((c) => {
                return (
                  <SelectPrimitive.Option
                    key={parseName(c)}
                    value={parseName(c)}
                  >
                    {parseName(c)}
                  </SelectPrimitive.Option>
                );
              })}
            </SelectPrimitive.Root>
          </div>
        </div>
        <div className="flex flex-row gap-md items-center">
          <div className="bodySm text-text-soft flex-1">
            All config entries be mounted on path specified in the container.
          </div>
          <Button
            content="Add Config Mount"
            variant="basic"
            disabled={!mountPath || !refName}
            onClick={() => {
              setConfigMounts((prev) => [...prev, { mountPath, refName }]);
            }}
          />
        </div>
      </div>
      {configMounts && configMounts.length > 0 && (
        <ConfigMountsList
          configMounts={configMounts}
          onDelete={(cm) => {
            setConfigMounts((prev) => prev.filter((p) => p !== cm));
          }}
        />
      )}
    </>
  );
};

const AppEnvironment = () => {
  const { envPage, setEnvPage, setPage } = useAppState();
  const [configMounts, setConfigMounts] = useState<Array<IConfigMount>>([]);
  const items: {
    label: string;
    value: createAppEnvPage;
  }[] = [
    {
      label: `Environment variables`,
      value: 'environment_variables',
    },
    {
      label: 'Config mount',
      value: 'config_mounts',
    },
  ];

  return (
    <FadeIn
      onSubmit={(e) => {
        e.preventDefault();
      }}
    >
      <div className="flex flex-col gap-xl ">
        <div className="headingXl text-text-default">Environment</div>
        <ExtendedFilledTab
          value={envPage}
          onChange={setEnvPage}
          items={items}
        />
      </div>
      <AnimatePresence mode="wait">
        <motion.div
          key={envPage || 'empty'}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="flex flex-col gap-6xl w-full"
        >
          {envPage === 'environment_variables' && <EnvironmentVariables />}
          {envPage === 'config_mounts' && (
            <ConfigMounts
              configMounts={configMounts}
              setConfigMounts={setConfigMounts}
            />
          )}
        </motion.div>
      </AnimatePresence>

      <div className="flex flex-row gap-xl justify-end items-center">
        <Button
          content="Save & Go Back"
          prefix={<ArrowLeft />}
          variant="outline"
          onClick={() => {
            setPage('compute');
          }}
        />

        <div className="text-surface-primary-subdued">|</div>

        <Button
          content="Save & Continue"
          suffix={<ArrowRight />}
          variant="primary"
          onClick={() => {
            setPage('network');
          }}
        />
      </div>
    </FadeIn>
  );
};

export default AppEnvironment;
