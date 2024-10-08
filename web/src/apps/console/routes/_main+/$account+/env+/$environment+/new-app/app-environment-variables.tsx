import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { Chip, ChipGroup } from '~/components/atoms/chips';
import { TextInput } from '~/components/atoms/input';
import Tooltip from '~/components/atoms/tooltip';
import { usePagination } from '~/components/molecule/pagination';
import { cn } from '~/components/utils';
import { listFlex } from '~/console/components/console-list-components';
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
import List from '~/console/components/list';
import NoResultsFound from '~/console/components/no-results-found';
import { IShowDialog } from '~/console/components/types.d';
import { useAppState } from '~/console/page-components/app-states';
import useForm from '~/root/lib/client/hooks/use-form';
import {
  DISCARD_ACTIONS,
  useUnsavedChanges,
} from '~/root/lib/client/hooks/use-unsaved-changes';
import Yup from '~/root/lib/server/helpers/yup';
import { NonNullableString } from '~/root/lib/types/common';
import AppDialog from './app-dialogs';

interface IEnvVariable {
  key: string;
  type?: 'config' | 'secret' | 'pvc' | NonNullableString;
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
  type: 'config' | 'secret' | 'pvc' | NonNullableString;
}

const EnvironmentVariablesList = ({
  envVariables,
  onDelete = (_) => _,
}: IEnvVariablesList) => {
  const { page, hasNext, hasPrevious, onNext, onPrev, setItems } =
    usePagination({
      items: envVariables || [],
      itemsPerPage: 5,
    });

  useEffect(() => {
    setItems(envVariables || []);
  }, [envVariables]);

  return (
    <div className="flex flex-col bg-surface-basic-default">
      {envVariables?.length > 0 && (
        <List.Root
          className="min-h-[347px] !shadow-none"
          header={
            <div className="flex flex-row items-center w-full">
              <div className="text-text-strong bodyMd flex-1">
                Environment variable list
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
                key={ev.key}
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
                    className: 'w-[80px]',
                    render: () => (
                      <Tooltip.Root
                        className="!max-w-[400px]"
                        content={
                          <div className="bodyMd-semibold text-text-default">
                            {ev.key}
                          </div>
                        }
                      >
                        <div className="bodyMd-semibold text-text-default truncate w-fit max-w-full">
                          <div className="truncate">
                            <span>{ev.key}</span>
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
                            {!ev.type && ev.value}
                            {!!ev.type && (
                              <>
                                <span className="line-clamp-1">
                                  {ev.refName}
                                </span>
                                <span>
                                  <ArrowRight size={16} />
                                </span>
                                {ev.refKey}
                              </>
                            )}
                          </div>
                        }
                      >
                        <div className="flex flex-row gap-md items-center bodyMd text-text-soft">
                          {!ev.type && ev.value}
                          {!!ev.type && (
                            <>
                              <span className="line-clamp-1">{ev.refName}</span>
                              <span>
                                <ArrowRight size={16} />
                              </span>
                              {ev.refKey}
                            </>
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
      {envVariables?.length === 0 && (
        <div className="rounded border-border-default border min-h-[347px] flex flex-row items-center justify-center">
          <NoResultsFound
            title={null}
            subtitle="No environment variables are added."
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

export const EnvironmentVariables = () => {
  const { setContainer, getContainer, getReadOnlyContainer, readOnlyApp } =
    useAppState();

  const [showCSDialog, setShowCSDialog] = useState<IShowDialog>(null);
  const { performAction } = useUnsavedChanges();

  const entry = Yup.object({
    type: Yup.string().oneOf(['config', 'secret']).notRequired(),
    key: Yup.string().required(),

    refKey: Yup.string()
      .when(['type'], ([type], schema) => {
        if (type === 'config' || type === 'secret') {
          return schema.required();
        }
        return schema;
      })
      .notRequired(),
    refName: Yup.string()
      .when(['type'], ([type], schema) => {
        if (type === 'config' || type === 'secret') {
          return schema.required();
        }
        return schema;
      })
      .notRequired(),
  });

  const {
    values,
    setValues,
    submit,
    resetValues: reset,
  } = useForm({
    initialValues: getReadOnlyContainer().env || null,
    validationSchema: Yup.array(entry),
    onSubmit: (val) => {
      // @ts-ignore
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
    const tempVal = val || [];
    setValues((v = []) => {
      const data = {
        key: tempVal.key,
        type: tempVal.type,
        refName: tempVal.refName || '',
        refKey: tempVal.refKey || '',
        value: tempVal.value || '',
      };
      return [...(v || []), data];
    });
  };

  const removeEntry = (val: IEnvVariable) => {
    // @ts-ignore
    setValues((v) => {
      const nv = v?.filter((v) => v.key !== val.key);
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
    handleSubmit,
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
          value: '',
        };

        // setEnvVariables((prev) => [...prev, ev]);
        addEntry(ev);
      }
      resetValues();
    },
  });

  useEffect(() => {
    if (performAction === DISCARD_ACTIONS.DISCARD_CHANGES) {
      // if (app.ciBuildId) {
      //   setIsEdited(false);
      // }
      reset();
      // @ts-ignore
      // setBuildData(readOnlyApp?.build);
    }

    // else if (performAction === 'init') {
    //   setIsEdited(false);
    // }
  }, [performAction]);

  useEffect(() => {
    reset();
  }, [readOnlyApp]);

  return (
    <>
      <form
        onSubmit={handleSubmit}
        className="flex flex-col gap-3xl p-3xl rounded border border-border-default"
      >
        <div className="flex flex-row gap-3xl items-start">
          <div className="basis-1/3">
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
                <div className="flex flex-row items-center rounded border border-border-default bg-surface-basic-default line-clamp-1">
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
                  <Tooltip.Root
                    className="!max-w-[400px]"
                    content={
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
                      <Chip
                        item={{ name: 'mres' }}
                        label="Managed resources"
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
            type="submit"
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
      </form>
      <EnvironmentVariablesList
        envVariables={getContainer().env || []}
        onDelete={(ev) => {
          removeEntry(ev);
          // setEnvVariables((prev) => prev.filter((p) => p !== ev));
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
