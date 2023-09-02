import {
  ArrowRight,
  LockSimple,
  LockSimpleOpen,
  X,
  XCircleFill,
} from '@jengaicons/react';
import { useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { Chip, ChipGroup } from '~/components/atoms/chips';
import { TextInput } from '~/components/atoms/input';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import List from '~/console/components/list';
import Select from '~/components/atoms/select';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { parseName, parseNodes } from '~/console/server/r-urils/common';
import { useParams } from '@remix-run/react';
import { AnimatePresence, motion } from 'framer-motion';
import { handleError } from '~/root/lib/utils/common';
import AppDialog, { IShow } from './app-dialogs';

interface IEnvVariable {
  type: 'config' | 'secrets' | 'literal';
  key: string;
  value: string | null;
  refName: string | null;
  refKey: string | null;
}

interface IEnvVariablesList {
  envVariables: Array<IEnvVariable>;
  onDelete: (envVariable: IEnvVariable) => void;
}

interface IEnvironmentVariables {
  envVariables: Array<IEnvVariable>;
  setEnvVariables: React.Dispatch<React.SetStateAction<Array<IEnvVariable>>>;
}

export interface IValue {
  variable: string;
  key: string;
  type: 'config' | 'secrets';
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
            <List.Item
              key={ev.key}
              items={[
                {
                  key: `${index}-column-0`,
                  render: () => (
                    <div className="text-icon-default">
                      {ev.type === 'config' && (
                        <LockSimpleOpen color="currentColor" size={16} />
                      )}
                      {ev.type === 'secrets' && (
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
                      {ev.type === 'literal' && ev.value}
                      {ev.type !== 'literal' && (
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

const EnvironmentVariables = ({
  envVariables,
  setEnvVariables,
}: IEnvironmentVariables) => {
  const [showCSDialog, setShowCSDialog] = useState<IShow>(null);
  const [textInputValue, setTextInputValue] = useState<string>('');
  const [value, setValue] = useState<IValue | null>(null);
  const [key, setKey] = useState<string>('');
  const [keyValueError, setKeyValueError] = useState<string | null>(null);
  return (
    <>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-default">
        <div className="flex flex-row gap-3xl items-center">
          <div className="flex-1">
            <TextInput
              label="Key"
              size="lg"
              error={!!keyValueError}
              message={keyValueError}
              value={key}
              autoComplete="off"
              onChange={({ target }) => {
                setKey(target.value);
              }}
            />
          </div>
          <div className="flex-1">
            {value ? (
              <div className="flex flex-col gap-md">
                <div className="bodyMd-medium text-text-default">Value</div>
                <div className="flex flex-row items-center rounded border border-border-default bg-surface-basic-defaul">
                  <span className="py-lg pl-lg pr-md text-text-default">
                    {value.type === 'config' ? (
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
                    {value.variable}
                    <ArrowRight size={16} weight={1} />
                    {value.key}
                  </div>
                  <button
                    tabIndex={-1}
                    type="button"
                    className="outline-none p-lg text-text-default rounded-full"
                    onClick={() => {
                      setValue(null);
                    }}
                  >
                    <XCircleFill size={16} color="currentColor" />
                  </button>
                </div>
              </div>
            ) : (
              <TextInput
                value={textInputValue}
                onChange={({ target }) => setTextInputValue(target.value)}
                label="Value"
                size="lg"
                suffix={
                  !textInputValue ? (
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
                showclear={!!textInputValue}
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
            disabled={!key || !(value || textInputValue)}
            onClick={() => {
              if (!envVariables.find((p) => p.key === key)) {
                if (textInputValue) {
                  const ev: IEnvVariable = {
                    key,
                    refKey: null,
                    refName: null,
                    type: 'literal',
                    value: textInputValue,
                  };
                  setEnvVariables((prev) => [...prev, ev]);
                  setTextInputValue('');
                } else {
                  if (value) {
                    const ev: IEnvVariable = {
                      key,
                      refKey: value.key,
                      refName: value.variable,
                      type: value.type,
                      value: null,
                    };

                    setEnvVariables((prev) => [...prev, ev]);
                    setValue(null);
                  }
                  setKey('');
                }
              } else {
                setKeyValueError(
                  'Key already exists in environment variables list.'
                );
              }
            }}
          />
        </div>
      </div>
      {envVariables && envVariables.length > 0 && (
        <EnvironmentVariablesList
          envVariables={envVariables}
          onDelete={(ev) => {
            setEnvVariables((prev) => prev.filter((p) => p !== ev));
          }}
        />
      )}
      <AppDialog
        show={showCSDialog}
        setShow={setShowCSDialog}
        onSubmit={(item) => {
          setValue(item);
          setShowCSDialog(false);
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
            <List.Item
              key={`${cm.mountPath} ${cm.refName}`}
              items={[
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
            <Select.Root
              disabled={isloading}
              label="Config"
              value={refName}
              onChange={({ target }: any) => {
                setRefName(target.value);
              }}
            >
              <Select.Option value="">--- Select config ---</Select.Option>
              {configs.map((c) => {
                return (
                  <Select.Option key={parseName(c)} value={parseName(c)}>
                    {parseName(c)}
                  </Select.Option>
                );
              })}
            </Select.Root>
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
  const [activeTab, setActiveTab] = useState('environment-variables');
  const [envVariables, setEnvVariables] = useState<Array<IEnvVariable>>([]);
  const [configMounts, setConfigMounts] = useState<Array<IConfigMount>>([]);

  return (
    <>
      <div className="flex flex-col gap-xl ">
        <div className="headingXl text-text-default">Environment</div>
        <ExtendedFilledTab
          value={activeTab}
          onChange={setActiveTab}
          items={[
            {
              label: `Environment variables`,
              value: `environment-variables`,
            },
            {
              label: 'Config mount',
              value: 'config-mount',
            },
          ]}
        />
      </div>
      <AnimatePresence mode="wait">
        <motion.div
          key={activeTab || 'empty'}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.1 }}
          className="flex flex-col gap-6xl w-full"
        >
          {activeTab === 'environment-variables' && (
            <EnvironmentVariables
              envVariables={envVariables}
              setEnvVariables={setEnvVariables}
            />
          )}
          {activeTab === 'config-mount' && (
            <ConfigMounts
              configMounts={configMounts}
              setConfigMounts={setConfigMounts}
            />
          )}
        </motion.div>
      </AnimatePresence>
    </>
  );
};

export default AppEnvironment;
