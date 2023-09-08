import { LockSimpleOpen, X } from '@jengaicons/react';
import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import { useParams } from '@remix-run/react';
import { handleError } from '~/root/lib/utils/common';
import { parseName, parseNodes } from '~/console/server/r-urils/common';
import { NonNullableString } from '~/root/lib/types/common';
import List from '~/console/components/list';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import SelectPrimitive from '~/components/atoms/select-primitive';
import { InfoLabel } from './util';
import { useAppState } from './states';

export interface IValue {
  refKey: string;
  refName: string;
  type: 'config' | 'secret' | NonNullableString;
}

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

export const ConfigMounts = () => {
  const api = useAPIClient();

  const [isloading, setIsloading] = useState<boolean>(true);
  const { workspace, project, scope } = useParams();
  const [configs, setConfigs] = useState([]);

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

  const { getContainer, setContainer } = useAppState();
  const { volumes } = getContainer();

  const { setValues, submit, values } = useForm({
    initialValues: volumes || [],
    validationSchema: Yup.array(
      Yup.object({
        mountPath: Yup.string().required(),
        refName: Yup.string().required(),
      })
    ),
    onSubmit: (val) => {
      setContainer((c) => ({
        ...c,
        volumes: val,
      }));
    },
  });

  const [mountPath, setMountPath] = useState('');

  type Ientry = {
    refName: string;
    mountPath: string;
  };

  const addEntry = (val: Ientry) => {
    setValues((v) => {
      return [
        ...v,
        {
          type: 'config',
          mountPath: val.mountPath,
          refName: val.refName,
        },
      ];
    });
  };

  const deleteEntry = (val: { mountPath: string }) => {
    setValues((v) => {
      return v.filter((v) => v.mountPath !== val.mountPath);
    });
  };

  const [entryError, setEntryError] = useState('');

  useEffect(() => {
    submit();
  }, [values]);

  useEffect(() => {
    setEntryError('');
  }, [mountPath]);

  return (
    <>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-default">
        <div className="flex flex-row gap-3xl items-center">
          <div className="flex-1">
            <TextInput
              label={<InfoLabel label="Path" info="some usefull information" />}
              size="lg"
              autoComplete="off"
              error={!!entryError}
              message={entryError}
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
              if (values.find((k) => k.mountPath === mountPath)) {
                setEntryError('path already present');
                return;
              }
              addEntry({ mountPath, refName });
              setMountPath('');
              setRefName('');
            }}
          />
        </div>
      </div>
      {volumes && volumes.length > 0 && (
        <ConfigMountsList
          configMounts={volumes}
          onDelete={(cm) => {
            // setConfigMounts((prev) => prev.filter((p) => p !== cm));
            deleteEntry(cm);
          }}
        />
      )}
    </>
  );
};
