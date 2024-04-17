import { useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import SelectPrimitive from '~/components/atoms/select-primitive';
import { usePagination } from '~/components/molecule/pagination';
import { cn } from '~/components/utils';
import { InfoLabel } from '~/iotconsole/components/commons';
import List from '~/iotconsole/components/list';
import NoResultsFound from '~/iotconsole/components/no-results-found';
import { useAppState } from '~/iotconsole/page-components/app-states';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/iotconsole/server/r-utils/common';
import useDebounce from '~/lib/client/hooks/use-debounce';
import useForm from '~/lib/client/hooks/use-form';
import { useUnsavedChanges } from '~/lib/client/hooks/use-unsaved-changes';
import Yup from '~/lib/server/helpers/yup';
import { NonNullableString } from '~/lib/types/common';
import { handleError } from '~/lib/utils/common';
import {
  ChevronLeft,
  ChevronRight,
  LockSimpleOpen,
  SmileySad,
  X,
} from '~/iotconsole/components/icons';
import { IConfigs } from '~/iotconsole/server/gql/queries/iot-config-queries';
import { Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret as ConfigOrSecretType } from '~/root/src/generated/gql/server';

export interface IValue {
  refKey: string;
  refName: string;
  type: ConfigOrSecretType | NonNullableString;
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
  const { page, hasNext, hasPrevious, onNext, onPrev, setItems } =
    usePagination({
      items: configMounts,
      itemsPerPage: 5,
    });

  useEffect(() => {
    setItems(configMounts);
  }, [configMounts]);

  return (
    <div className="flex flex-col gap-lg bg-surface-basic-default">
      {configMounts.length > 0 && (
        <List.Root
          className="min-h-[347px] !shadow-none"
          header={
            <div className="flex flex-row items-center">
              <div className="text-text-strong bodyMd flex-1">
                Config mount list
              </div>
              <div className="flex flex-row items-center">
                <IconButton
                  icon={<ChevronLeft />}
                  size="xs"
                  variant="plain"
                  onClick={() => onPrev()}
                  disabled={!hasPrevious}
                />
                <IconButton
                  icon={<ChevronRight />}
                  size="xs"
                  variant="plain"
                  onClick={() => onNext()}
                  disabled={!hasNext}
                />
              </div>
            </div>
          }
        >
          {page.map((cm, index) => {
            return (
              <List.Row
                className={cn({
                  '!border-b': index < 4,
                  '!rounded-b-none': index < 4,
                })}
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
      )}
      {configMounts.length === 0 && (
        <div className="rounded border-border-default border min-h-[347px] flex flex-row items-center justify-center">
          <NoResultsFound
            title={null}
            subtitle="No config mounts are added."
            compact
            image={<SmileySad size={32} weight={1} />}
            shadow={false}
            border={false}
          />
        </div>
      )}
    </div>
  );
};

export const ConfigMounts = () => {
  const api = useIotConsoleApi();
  const [isloading, setIsloading] = useState<boolean>(true);
  const { environment, project } = useParams();
  const [configs, setConfigs] = useState<ExtractNodeType<IConfigs>[]>([]);

  const [refName, setRefName] = useState<string>('');

  useDebounce(
    async () => {
      if (!project || !environment) {
        throw new Error('Project and Environment is requireed!.');
      }
      try {
        setIsloading(true);
        const { data, errors } = await api.listConfigs({
          envName: environment,
          projectName: project,
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

  // for updating
  const { hasChanges } = useUnsavedChanges();

  const { setValues, submit, values, resetValues } = useForm({
    initialValues: volumes,
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
        ...(v || []),
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
      return v?.filter((v) => v.mountPath !== val.mountPath);
    });
  };

  const [entryError, setEntryError] = useState('');

  useEffect(() => {
    submit();
  }, [values]);

  useEffect(() => {
    setEntryError('');
  }, [mountPath]);

  // for updating
  useEffect(() => {
    if (!hasChanges) {
      resetValues();
    }
  }, [hasChanges]);

  return (
    <>
      <form
        onSubmit={(e) => {
          e.preventDefault();

          if (values?.find((k) => k.mountPath === mountPath)) {
            setEntryError('path already present');
            return;
          }
          addEntry({ mountPath, refName });
          setMountPath('');
          setRefName('');
        }}
        className="flex flex-col gap-3xl p-3xl rounded border border-border-default"
      >
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
            type="submit"
            content="Add Config Mount"
            variant="basic"
            disabled={!mountPath || !refName}
            // onClick={() => {
            //
            // }}
          />
        </div>
      </form>
      <ConfigMountsList
        configMounts={volumes || []}
        onDelete={(cm) => {
          deleteEntry(cm);
        }}
      />
    </>
  );
};
