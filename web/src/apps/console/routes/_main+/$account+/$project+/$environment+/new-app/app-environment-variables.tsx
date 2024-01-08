import {
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  LockSimple,
  LockSimpleOpen,
  SmileySad,
  X,
  XCircleFill,
} from '@jengaicons/react';
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
import AppDialog from './app-dialogs';

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
  const { page, hasNext, hasPrevious, onNext, onPrev, setItems } =
    usePagination({
      items: envVariables,
      itemsPerPage: 5,
    });

  useEffect(() => {
    setItems(envVariables);
  }, [envVariables]);

  return (
    <div className="flex flex-col bg-surface-basic-default">
      {envVariables.length > 0 && (
        <List.Root
          className="min-h-[347px] !shadow-none"
          header={
            <div className="flex flex-row items-center">
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
                        {!ev.type && ev.value}
                        {!!ev.type && (
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
      )}
      {envVariables.length === 0 && (
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
  const { setContainer, getContainer } = useAppState();

  const [showCSDialog, setShowCSDialog] = useState<IShowDialog>(null);

  const entry = Yup.object({
    type: Yup.string().oneOf(['config', 'secret']).notRequired(),
    key: Yup.string().required(),

    value: Yup.string().when(['type'], ([type], schema) => {
      if (type === undefined) {
        return schema.required();
      }
      return schema;
    }),
    refKey: Yup.string()
      .when(['type'], ([type], schema) => {
        if (type === 'config' || type === 'secret') {
          // console.log('here', type);
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
      // console.log(v, 'here');
      return schema;
    }),
    refName: Yup.string().when('$textInputValue', ([v], schema) => {
      // console.log(v);
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
      <form
        onSubmit={handleSubmit}
        className="flex flex-col gap-3xl p-3xl rounded border border-border-default"
      >
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
