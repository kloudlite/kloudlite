import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { Chip, ChipGroup } from '~/components/atoms/chips';
import { TextInput } from '~/components/atoms/input';
import { usePagination } from '~/components/molecule/pagination';
import { cn } from '~/components/utils';
import List from '~/console/components/list';
import NoResultsFound from '~/console/components/no-results-found';
import { IShowDialog } from '~/console/components/types.d';
import { useAppState } from '~/console/page-components/app-states';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { NonNullableString } from '~/root/lib/types/common';
import {
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  LockSimple,
  LockSimpleOpen,
  SmileySad,
  X,
  XCircleFill,
} from '~/console/components/icons';
import Tooltip from '~/components/atoms/tooltip';
import { listFlex } from '~/console/components/console-list-components';
import AppDialog from './app-dialogs';

interface IConfigMount {
  type: 'config' | 'secret' | 'pvc';
  refName: string;
  mountPath: string;
}

interface IConfigMountList {
  configMounts: Array<IConfigMount>;
  onDelete: (configMount: IConfigMount) => void;
}

export interface IValue {
  refKey: string;
  refName: string;
  type: 'config' | 'secret' | 'pvc' | NonNullableString;
}

const ConfigMountList = ({
  configMounts,
  onDelete = (_) => _,
}: IConfigMountList) => {
  const { page, hasNext, hasPrevious, onNext, onPrev, setItems } =
    usePagination({
      items: configMounts || [],
      itemsPerPage: 5,
    });

  useEffect(() => {
    setItems(configMounts || []);
  }, [configMounts]);

  return (
    <div className="flex flex-col bg-surface-basic-default">
      {configMounts?.length > 0 && (
        <List.Root
          className="min-h-[347px] !shadow-none"
          header={
            <div className="flex flex-row items-center w-full">
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
          {page.map((ev, index) => {
            return (
              <List.Row
                key={ev.mountPath}
                className={cn({
                  '!border-b': index < 4,
                  '!rounded-b-none': index < 4,
                })}
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
                    className: 'w-[80px] truncate',
                    render: () => (
                      <Tooltip.Root
                        className="!max-w-[400px]"
                        content={
                          <div className="bodyMd-semibold text-text-default truncate">
                            {ev.mountPath}
                          </div>
                        }
                      >
                        <div className="bodyMd-semibold text-text-default truncate w-fit max-w-full">
                          <div className="truncate">
                            <span>{ev.mountPath}</span>
                          </div>
                        </div>
                      </Tooltip.Root>
                    ),
                  },
                  {
                    key: `${index}-column-2`,
                    render: () => (
                      <Tooltip.Root
                        className="!max-w-[400px]"
                        content={
                          <div className="flex flex-row gap-md items-center bodyMd text-text-soft">
                            {!!ev.type && (
                              <span className="line-clamp-1">{ev.refName}</span>
                            )}
                          </div>
                        }
                      >
                        <div className="flex flex-row gap-md items-center bodyMd text-text-soft">
                          {!!ev.type && (
                            <span className="line-clamp-1">{ev.refName}</span>
                          )}
                        </div>
                      </Tooltip.Root>
                    ),
                  },
                  listFlex({ key: 'flex-1' }),
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
      )}
      {configMounts?.length === 0 && (
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
  const { setContainer, getContainer } = useAppState();

  const [showCSDialog, setShowCSDialog] = useState<IShowDialog>(null);

  // for updating

  const entry = Yup.object({
    type: Yup.string().oneOf(['config', 'secret']).notRequired(),
    refName: Yup.string()
      .when(['type'], ([type], schema) => {
        if (type === 'config' || type === 'secret') {
          return schema.required();
        }
        return schema;
      })
      .notRequired(),
    mountPath: Yup.string().required(),
  });

  const { values, setValues, submit } = useForm({
    initialValues: getContainer().volumes,
    validationSchema: Yup.array(entry),
    onSubmit: (val) => {
      setContainer((c) => ({
        ...c,
        volumes: val,
      }));
    },
  });

  useEffect(() => {
    submit();
  }, [values]);

  const addEntry = (val: IConfigMount) => {
    const tempVal = val || [];
    setValues((v = []) => {
      const data = {
        type: tempVal.type,
        refName: tempVal.refName,
        mountPath: tempVal.mountPath,
      };
      return [...(v || []), data];
    });
  };

  const removeEntry = (val: IConfigMount) => {
    setValues((v) => {
      const nv = v?.filter((v) => v.mountPath !== val.mountPath);
      return nv;
    });
  };

  const vSchema = Yup.object({
    refKey: Yup.string().when('$textInputValue', ([_v], schema) => {
      return schema;
    }),
    refName: Yup.string().when('$textInputValue', ([_v], schema) => {
      return schema;
    }),
  });

  interface InitialValuesProps {
    mountPath: string;
    value?: IValue;
  }

  const initialValues: InitialValuesProps = {
    mountPath: '',
  };

  const {
    values: eValues,
    errors: eErrors,
    handleChange: eHandleChange,
    handleSubmit,
    setValues: eSetValues,
    resetValues,
    submit: eSubmit,
  } = useForm({
    initialValues,
    validationSchema: Yup.object().shape({
      value: vSchema,
      textInputValue: Yup.string(),
      mountPath: Yup.string()
        .required()
        .test('is-valid', 'Path already exists.', (value) => {
          return !getContainer().volumes?.find((v) => v.mountPath === value);
        }),
    }),
    onSubmit: () => {
      if (eValues.value) {
        const ev: IConfigMount = {
          refName: eValues.value.refName,
          type: eValues.value.type,
          mountPath: eValues.mountPath,
        };
        addEntry(ev);
      }
      resetValues();
    },
  });

  return (
    <>
      <form
        onSubmit={handleSubmit}
        className="flex flex-col gap-3xl p-3xl rounded border border-border-default"
      >
        <div className="flex flex-row gap-3xl items-start">
          <div className="basis-1/3">
            <TextInput
              label="Path"
              size="lg"
              error={!!eErrors.mountPath}
              message={eErrors.mountPath}
              value={eValues.mountPath}
              autoComplete="off"
              onChange={eHandleChange('mountPath')}
            />
          </div>
          <div className="flex-1 ">
            <div className="flex flex-col">
              <div className="bodyMd-medium text-text-default pb-md">
                <span className="h-4xl block">Value</span>
              </div>
              <div className="h-[43px] px-lg flex flex-row items-center rounded border border-border-default bg-surface-basic-default line-clamp-1">
                {eValues.value ? (
                  <div className="flex flex-row items-center">
                    <span className="py-lg pl-lg pr-md text-text-default">
                      {eValues.value.type === 'config' ? (
                        <LockSimpleOpen
                          size={14}
                          weight={2.5}
                          color="currentColor"
                        />
                      ) : (
                        <LockSimple
                          size={14}
                          weight={2.5}
                          color="currentColor"
                        />
                      )}
                    </span>
                    <Tooltip.Root
                      className="!max-w-[400px]"
                      content={
                        <div className="flex-1 flex flex-row gap-md items-center py-xl px-lg bodyMd text-text-soft ">
                          <span>{eValues.value.refName}</span>
                          <span>
                            <ArrowRight size={16} />
                          </span>
                          <span className="truncate">
                            {eValues.value.refKey}
                          </span>
                        </div>
                      }
                    >
                      <div className="flex-1 flex flex-row gap-md items-center py-xl px-lg bodyMd text-text-soft ">
                        <span className="line-clamp-1">
                          {eValues.value.refName}
                        </span>
                        <span>
                          <ArrowRight size={16} />
                        </span>

                        <span className="line-clamp-1">
                          {eValues.value.refKey}
                        </span>
                      </div>
                    </Tooltip.Root>
                    <button
                      aria-label="clear"
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
                ) : (
                  <div className="w-full flex flex-row justify-between">
                    <div />
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
                      <Chip
                        item={{ name: 'mres' }}
                        label="Integrated resources"
                        type="CLICKABLE"
                      />
                    </ChipGroup>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
        <div className="flex flex-row gap-md items-center">
          <div className="bodySm text-text-soft flex-1">
            All config entries be mounted on path specified in the container.
          </div>
          <Button
            type="submit"
            content="Add config mount"
            variant="basic"
            disabled={!eValues.mountPath || !eValues.value}
            onClick={() => {
              eSubmit();
            }}
          />
        </div>
      </form>
      <ConfigMountList
        configMounts={getContainer().volumes || []}
        onDelete={(ev) => {
          removeEntry(ev);
        }}
      />
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
