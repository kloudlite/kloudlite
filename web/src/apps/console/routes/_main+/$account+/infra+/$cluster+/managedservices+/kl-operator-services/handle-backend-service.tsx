/* eslint-disable guard-for-in */
/* eslint-disable react/destructuring-assignment */
import { ArrowRight, Search, Check } from '@jengaicons/react';
import { useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import ActionList from '~/components/atoms/action-list';
import { NumberInput, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { ListTitle } from '~/console/components/console-list-components';
import { IdSelector } from '~/console/components/id-selector';
import List from '~/console/components/list';
import NoResultsFound from '~/console/components/no-results-found';
import { IDialogBase } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IClusterMSvs } from '~/console/server/gql/queries/cluster-managed-services-queries';
import { ExtractNodeType } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useInputSearch } from '~/root/lib/client/helpers/search-filter';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { NN } from '~/root/lib/types/common';
import { handleError } from '~/root/lib/utils/common';
import { IMSvTemplates } from '~/console/server/gql/queries/managed-templates-queries';
import { Switch } from '~/components/atoms/switch';
import { cn } from '~/components/utils';

type IDialog = IDialogBase<ExtractNodeType<IClusterMSvs>> & {
  templates: IMSvTemplates;
};

type IActiveCategory = {
  name: string;
  displayName: string;
} | null;

type ISelectedService = {
  category: {
    name: string;
    displayName: string;
  };

  service: NN<IMSvTemplates>[number]['items'][number];
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
  templates: IMSvTemplates;
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
          <NoResultsFound
            shadow={false}
            border={false}
            title={<div className="pt-2xl" />}
            subtitle="No Services Available now"
          />
        ) : (
          <List.Root>
            {searchResults.map((item) => {
              return (
                <List.Row
                  className={cn('group/team')}
                  pressed={selectedService?.service.name === item.name}
                  key={`${item.name}row`}
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
                  columns={[
                    {
                      key: item.name,
                      className: 'flex-grow',
                      render: () => (
                        <ListTitle
                          avatar={
                            <img
                              width={25}
                              src={item.logoUrl}
                              alt={item.displayName}
                            />
                          }
                          title={
                            <div className="bodySm">{item.displayName}</div>
                          }
                        />
                      ),
                    },
                    {
                      key: `${item.name}-arrow`,
                      render: () =>
                        selectedService?.service.name === item.name ? (
                          <Check size={16} />
                        ) : (
                          <div className="invisible transition-all delay-100 duration-10 group-hover/team:visible group-hover/team:translate-x-sm">
                            <ArrowRight size={24} />
                          </div>
                        ),
                    },
                  ]}
                />
              );
            })}
          </List.Root>
        )}

        {/* <Grid.Root className="!gap-4xl">
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

                          <div>
                            <div
                              key={item.name}
                              className={cn(
                                'bodySm-semibold text-text-default line-clamp-2 text-center',
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
        </Grid.Root> */}
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
  onChange: (e: string) => (e: { target: { value: any } }) => void;
  value: any;
  error: boolean;
  message?: string;
}) => {
  console.log('value', value);

  const [qos, setQos] = useState(false);
  if (field.inputType === 'Number') {
    return (
      <NumberInput
        error={error}
        message={message}
        label={`${field.label}${field.required ? ' *' : ''}`}
        min={field.min}
        max={field.max}
        placeholder={field.label}
        value={parseFloat(value) / (field.multiplier || 1)}
        onChange={({ target }) => {
          onChange(`${field.name}`)(
            dummyEvent(
              `${parseFloat(target.value) * (field.multiplier || 1)}${
                field.unit
              }`
            )
          );
        }}
        suffix={field.displayUnit}
      />
    );
  }

  if (field.inputType === 'String') {
    return (
      <TextInput
        label={field.label}
        value={value}
        onChange={onChange(`${field.name}`)}
        suffix={field.displayUnit}
      />
    );
  }
  if (field.inputType === 'Resource') {
    return (
      <div className="flex flex-col gap-md">
        <div className="bodyMd-medium text-text-default">{`${field.label}${
          field.required ? ' *' : ''
        }`}</div>
        <div className="flex flex-row gap-xl items-center">
          <div className="flex flex-row gap-xl items-end flex-1 ">
            <div className="flex-1">
              <NumberInput
                error={error}
                message={message}
                min={field.min}
                max={field.max}
                placeholder={qos ? field.label : `${field.label} min`}
                value={parseFloat(value.min) / (field.multiplier || 1)}
                onChange={({ target }) => {
                  onChange(`${field.name}.min`)(
                    dummyEvent(
                      `${parseFloat(target.value) * (field.multiplier || 1)}${
                        field.unit
                      }`
                    )
                  );
                  if (qos) {
                    onChange(`${field.name}.max`)(
                      dummyEvent(
                        `${parseFloat(target.value) * (field.multiplier || 1)}${
                          field.unit
                        }`
                      )
                    );
                  }
                }}
                suffix={field.displayUnit}
              />
            </div>
            {!qos && (
              <div className="flex-1">
                <NumberInput
                  error={error}
                  message={message}
                  min={field.min}
                  max={field.max}
                  placeholder={`${field.label} max`}
                  value={parseFloat(value.max) / (field.multiplier || 1)}
                  onChange={({ target }) => {
                    onChange(`${field.name}.max`)(
                      dummyEvent(
                        `${parseFloat(target.value) * (field.multiplier || 1)}${
                          field.unit
                        }`
                      )
                    );
                  }}
                  suffix={field.displayUnit}
                />
              </div>
            )}
          </div>
          <div className="flex flex-col gap-md min-w-[115px]">
            <Switch
              label="Guaranteed"
              checked={qos}
              onChange={(_value) => {
                setQos(_value);
                if (_value) {
                  onChange(`${field.name}.max`)(dummyEvent(`${value.min}`));
                }
              }}
            />
          </div>
        </div>
      </div>
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
  console.log('values: ', values);

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
        resType="managed_resource"
        onChange={(v) => {
          handleChange('name')(dummyEvent(v));
        }}
      />
      {selectedService?.service.fields.map((field) => {
        const k = field.name;
        const x = k.split('.').reduce((acc, curr) => {
          if (!acc) {
            return values[curr];
          }
          return acc[curr];
        }, null);

        return (
          <RenderField
            error={!!errors[k]}
            message={errors[k]}
            value={x}
            onChange={handleChange}
            key={k}
            field={field}
          />
        );
      })}
    </div>
  );
};

const flatM = (obj) => {
  const flatJson = {};
  for (const key in obj) {
    const parts = key.split('.');
    let temp = flatJson;

    parts.forEach((part, index) => {
      if (index === parts.length - 1) {
        temp[part] = {
          min: null,
          max: null,
        };
      } else {
        temp[part] = temp[part] || {};
      }
      temp = temp[part];
    });
  }
  console.log('flatjson ', flatJson);

  return flatJson;
};

const Root = (props: IDialog) => {
  const { isUpdate, setVisible, templates } = props;
  const [activeCategory, setActiveCategory] = useState<IActiveCategory>(null);
  const [selectedService, setSelectedService] =
    useState<ISelectedService>(null);

  const api = useConsoleApi();
  const reload = useReload();

  const { cluster } = useParams();
  const [step, setStep] = useState<'choose' | 'fill'>('choose');

  console.log(selectedService);

  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    resetValues,
    isLoading,
    setValues,
  } = useForm({
    initialValues: {
      name: '',
      displayName: '',
    },
    validationSchema: Yup.object({}),
    onSubmit: async (val) => {
      const tempVal = { ...val };
      delete tempVal.name;
      delete tempVal.displayName;

      try {
        if (!cluster) {
          throw new Error('Cluster not found.');
        }
        if (
          !selectedService?.service.apiVersion ||
          !selectedService.service.kind
        ) {
          throw new Error('Service apiversion or kind error.');
        }
        const { errors: e } = await api.createClusterMSv({
          clusterName: cluster,
          service: {
            displayName: val.displayName,
            metadata: {
              name: val.name,
            },

            spec: {
              msvcSpec: {
                serviceTemplate: {
                  apiVersion: selectedService.service.apiVersion,
                  kind: selectedService.service.kind,
                  spec: {
                    ...tempVal,
                  },
                },
              },
              namespace: '',
            },
          },
        });
        if (e) {
          throw e[0];
        }
        setVisible(false);
        reload();
      } catch (err) {
        handleError(err);
      }
    },
  });

  useEffect(() => {
    if (selectedService?.service.fields) {
      setValues({
        name: '',
        displayName: '',
        ...flatM(
          selectedService?.service.fields.reduce((acc, curr) => {
            return { ...acc, [curr.name]: curr.defaultValue };
          }, {})
        ),
      });
    }
  }, [selectedService]);

  return (
    // <Popup.Header>
    //   {step === 'choose' ? (
    //     <div>Choose a service</div>
    //   ) : (
    //     <div className="flex flex-row items-center gap-2xl">
    //       <div className="flex flex-row items-center gap-lg">
    //         <IconButton
    //           icon={<ArrowLeft />}
    //           size="xs"
    //           variant="plain"
    //           onClick={() => {
    //             resetValues({});
    //             setStep('choose');
    //           }}
    //         />
    //         <img
    //           className="w-3xl h-3xl aspect-square"
    //           alt={selectedService?.service.displayName}
    //           src={selectedService?.service.logoUrl}
    //         />
    //       </div>
    //       <div>{selectedService?.service.displayName}</div>
    //     </div>
    //   )}
    // </Popup.Header>
    <Popup.Form
      onSubmit={(e) => {
        if (step === 'choose') {
          setStep('fill');
          e.preventDefault();
        } else handleSubmit(e);
      }}
    >
      <Popup.Content className="!min-h-[500px] !max-h-[500px]">
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
              setSelectedService(null);
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
    </Popup.Form>
  );
};

const HandleBackendService = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;
  return (
    <Popup.Root
      className="!min-w-[900px]"
      show={visible}
      onOpenChange={(v) => setVisible(v)}
    >
      <Popup.Header>
        {isUpdate ? 'Edit cloud provider' : 'Add new cloud provider'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleBackendService;
