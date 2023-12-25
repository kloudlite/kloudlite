import { ArrowLeft, Search } from '@jengaicons/react';
import { useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import ActionList from '~/components/atoms/action-list';
import { IconButton } from '~/components/atoms/button';
import { NumberInput, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { cn } from '~/components/utils';
import Grid from '~/console/components/grid';
import { IdSelector } from '~/console/components/id-selector';
import NoResultsFound from '~/console/components/no-results-found';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IManagedServiceTemplates } from '~/console/server/gql/queries/managed-service-queries';
import { parseTargetNs } from '~/console/server/r-utils/common';
import { useInputSearch } from '~/root/lib/client/helpers/search-filter';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { NN } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { IWorkspaceContext } from '../_.$account.$cluster.$project.$scope.$workspace/route';

type IActiveCategory = {
  name: string;
  displayName: string;
} | null;

type ISelectedService = {
  category: {
    name: string;
    displayName: string;
  };

  service: NN<IManagedServiceTemplates>[number]['items'][number];
} | null;

const ServicePicker = ({
  activeCategory,
  setActiveCategory,
  templates,
  selectedService,
  setSelectedService,
}: {
  activeCategory: IActiveCategory;
  setActiveCategory: React.Dispatch<IActiveCategory>;
  templates: IManagedServiceTemplates;
  selectedService: ISelectedService;
  setSelectedService: React.Dispatch<ISelectedService>;
}) => {
  const [searchProps, searchResults] = useInputSearch(
    {
      data:
        templates?.find((t) => t.category === activeCategory?.name)?.items ||
        [],
      keys: ['name'],
      reverse: false,
      threshold: 0.4,
    },
    [activeCategory]
  );
  return (
    <div className="flex min-h-[30vh]">
      <div className="pr-3xl min-w-[180px]">
        <ActionList.Root
          value={activeCategory?.name || ''}
          onChange={(v) => {
            setActiveCategory({
              name: v,
              displayName:
                templates?.find((t) => t.category === v)?.displayName || '',
            });
          }}
        >
          {templates?.map((t, index) => {
            if (!activeCategory && index === 0) {
              setActiveCategory({
                name: t.category,
                displayName: t.displayName,
              });
            }
            return (
              <ActionList.Item key={t.category} value={t.category}>
                {t.displayName}
              </ActionList.Item>
            );
          })}
        </ActionList.Root>
      </div>
      <div className="flex-1 pl-3xl flex flex-col gap-4xl">
        <TextInput
          {...searchProps}
          prefixIcon={<Search />}
          placeholder="Search"
        />
        {templates?.find((t) => t.category === activeCategory?.name)?.items
          .length === 0 ? (
          <NoResultsFound title="No Services Available now" />
        ) : null}
        <Grid.Root className="!gap-4xl">
          {searchResults.map((item) => {
            return (
              <Grid.Column
                onClick={() => {
                  if (!item.apiVersion) {
                    toast.error('not available now');
                    return;
                  }
                  if (activeCategory) {
                    setSelectedService({
                      category: activeCategory,
                      service: item,
                    });
                  }
                }}
                key={item.name}
                className={cn({
                  '!bg-surface-basic-active':
                    item.name === selectedService?.service.name,
                  'opacity-50 cursor-not-allowed': !item.apiVersion,
                })}
                rows={[
                  {
                    key: `${item.name}first`,
                    render() {
                      return (
                        <div className="flex flex-col gap-xl p-md overflow-hidden">
                          <img
                            className="w-5xl h-5xl aspect-square self-center"
                            src={item.logoUrl}
                            alt={item.displayName}
                          />

                          <div className="truncate overflow-hidden">
                            <div
                              key={item.name}
                              className={cn(
                                'bodyMd text-text-default truncate',
                                {
                                  'text-text-primary bodyMd-medium':
                                    item.name === selectedService?.service.name,
                                  'text-text-default':
                                    item.name !== selectedService?.service.name,
                                }
                              )}
                            >
                              {item.displayName}
                            </div>
                          </div>
                        </div>
                      );
                    },
                  },
                ]}
              />
            );
          })}
        </Grid.Root>
      </div>
    </div>
  );
};

const RenderField = ({
  field,
  value,
  onChange,
  error,
  message,
}: {
  field: NN<ISelectedService>['service']['fields'][number];
  onChange: (e: { target: { value: any } }) => void;
  value: any;
  error: boolean;
  message?: string;
}) => {
  if (field.inputType === 'Number') {
    return (
      <NumberInput
        error={error}
        message={message}
        label={`${field.label}${field.required ? ' *' : ''}`}
        min={field.min}
        max={field.max}
        placeholder={field.label}
        value={value}
        onChange={onChange}
        suffix={field.unit}
      />
    );
  }

  if (field.inputType === 'String') {
    return (
      <TextInput
        label={field.label}
        value={value}
        onChange={onChange}
        suffix={field.unit}
      />
    );
  }
  return <div>unknown input type {field.inputType}</div>;
};

const Fill = ({
  selectedService,
  values,
  handleChange,
  errors,
}: {
  selectedService: ISelectedService;
  values: { [key: string]: any };
  handleChange: (key: string) => (e: { target: { value: any } }) => void;
  errors: {
    [key: string]: string | undefined;
  };
}) => {
  return (
    <div className="flex flex-col gap-3xl min-h-[30vh]">
      <TextInput
        label="Name"
        value={values.displayName}
        onChange={handleChange('displayName')}
        error={!!errors.displayName}
        message={errors.displayName}
      />
      <IdSelector
        name={values.displayName}
        resType="managed_service"
        onChange={(v) => {
          handleChange('name')(dummyEvent(v));
        }}
      />
      {selectedService?.service.fields.map((field) => {
        const k = field.name;
        return (
          <RenderField
            error={!!errors[field.name]}
            message={errors[field.name]}
            value={values[k]}
            onChange={handleChange(field.name)}
            key={field.name}
            field={field}
          />
        );
      })}
    </div>
  );
};

const HandleBackendService = ({
  show,
  setShow,
  templates,
}: IDialog & {
  templates: IManagedServiceTemplates;
}) => {
  const [activeCategory, setActiveCategory] = useState<IActiveCategory>(null);
  const [selectedService, setSelectedService] =
    useState<ISelectedService>(null);

  const [step, setStep] = useState<'choose' | 'fill'>('choose');

  const api = useConsoleApi();
  const { workspace } = useOutletContext<IWorkspaceContext>();
  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: {
        name: '',
        displayName: '',
        ...(selectedService?.service.fields.reduce((acc, field) => {
          return {
            ...acc,
            [field.name]: field.defaultValue,
          };
        }, {}) || {}),
      } as {
        [key: string]: any;
      },
      validationSchema: Yup.object({
        name: Yup.string().required(),
        displayName: Yup.string().required(),
        ...selectedService?.service.fields.reduce((acc, curr) => {
          return {
            ...acc,
            [curr.name]: (() => {
              let returnYup: any = Yup;
              switch (curr.inputType) {
                case 'Number':
                  returnYup = returnYup.number();
                  if (curr.min) returnYup = returnYup.min(curr.min);
                  if (curr.max) returnYup = returnYup.max(curr.max);
                  break;
                case 'String':
                  returnYup = returnYup.string();
                  break;
                default:
                  toast.error(
                    `Unknown input type ${curr.inputType} for field ${curr.name}`
                  );
                  returnYup = returnYup.string();
              }

              if (curr.required) {
                returnYup = returnYup.required();
              }

              return returnYup;
            })(),
          };
        }, {}),
      }),
      onSubmit: async (val) => {
        const tempVal = { ...val };
        delete tempVal.name;
        delete tempVal.displayName;
        // try {
        //   const { errors: e } = await api.createManagedService({
        //     msvc: {
        //       displayName: val.displayName,
        //       metadata: {
        //         name: val.name,
        //         namespace: parseTargetNs(workspace),
        //         annotations: {},
        //       },
        //       spec: {
        //         msvcKind: {
        //           apiVersion: selectedService?.service.apiVersion || '',
        //           kind: selectedService?.service.kind || '',
        //         },
        //         inputs: {
        //           ...tempVal,
        //         },
        //       },
        //     },
        //   });
        //   if (e) {
        //     throw e[0];
        //   }
        // } catch (err) {
        //   handleError(err);
        // }
      },
    });

  return (
    <Popup.Root
      className="min-w-[800px] max-w-[850px] w-full"
      show={show as any}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }
        setShow(e);
      }}
    >
      <Popup.Header>
        {step === 'choose' ? (
          <div>Choose a service</div>
        ) : (
          <div className="flex flex-row items-center gap-2xl">
            <div className="flex flex-row items-center gap-lg">
              <IconButton
                icon={<ArrowLeft />}
                size="xs"
                variant="plain"
                onClick={() => {
                  resetValues({});
                  setStep('choose');
                }}
              />
              <img
                className="w-3xl h-3xl aspect-square"
                alt={selectedService?.service.displayName}
                src={selectedService?.service.logoUrl}
              />
            </div>
            <div>{selectedService?.service.displayName}</div>
          </div>
        )}
      </Popup.Header>
      <form
        onSubmit={(e) => {
          if (step === 'choose') {
            setStep('fill');
            e.preventDefault();
          } else handleSubmit(e);
        }}
      >
        <Popup.Content>
          {step === 'choose' ? (
            <ServicePicker
              {...{
                activeCategory,
                selectedService,
                setActiveCategory,
                setSelectedService,
                templates,
              }}
            />
          ) : (
            <Fill
              {...{
                templates,
                selectedService,
                values,
                errors,
                handleChange,
              }}
            />
          )}
        </Popup.Content>
        <Popup.Footer>
          {step === 'fill' ? (
            <Popup.Button
              type="button"
              variant="basic"
              onClick={() => {
                resetValues({});
                setStep('choose');
              }}
              content="Back"
            />
          ) : null}
          <Popup.Button
            disabled={!selectedService}
            loading={isLoading}
            type="submit"
            content={step === 'choose' ? 'Next' : 'Create'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleBackendService;
