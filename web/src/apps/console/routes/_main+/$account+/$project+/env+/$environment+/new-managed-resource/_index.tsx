/* eslint-disable guard-for-in */

import {
  useLoaderData,
  useNavigate,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import Select from '~/components/atoms/select';
import { NameIdView } from '~/console/components/name-id-view';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useForm, { dummyEvent } from '~/lib/client/hooks/use-form';
import Yup from '~/lib/server/helpers/yup';
import { FormEventHandler, useEffect, useState } from 'react';
import { IMSvTemplate } from '~/console/server/gql/queries/managed-templates-queries';
import { Switch } from '~/components/atoms/switch';
import { NumberInput, TextInput } from '~/components/atoms/input';
import { handleError } from '~/lib/utils/common';
import { titleCase, useMapper } from '~/components/utils';
import { IRemixCtx } from '~/lib/types/common';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { defer } from '@remix-run/node';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import { IProjectMSvs } from '~/console/server/gql/queries/project-managed-services-queries';
import { getManagedTemplate } from '~/console/utils/commons';
import MultiStepProgressWrapper from '~/console/components/multi-step-progress-wrapper';
import MultiStepProgress, {
  useMultiStepProgress,
} from '~/console/components/multi-step-progress';
import {
  BottomNavigation,
  ReviewComponent,
} from '~/console/components/commons';
import { IProjectContext } from '../../../_layout';

export const loader = (ctx: IRemixCtx) => {
  const { project } = ctx.params;
  const promise = pWrapper(async () => {
    const { data: mData, errors: mErrors } = await GQLServerHandler(
      ctx.request
    ).listProjectMSvs({
      projectName: project,
    });

    if (mErrors) {
      throw mErrors[0];
    }
    return { managedServicesData: mData };
  });
  return defer({ promise });
};

const RenderField = ({
  field,
  value,
  onChange,
  error,
  message,
}: {
  field: IMSvTemplate['fields'][number];
  onChange: (e: string) => (e: { target: { value: any } }) => void;
  value: any;
  error: boolean;
  message?: string;
}) => {
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
        value={parseFloat(value) / (field.multiplier || 1) || ''}
        onChange={({ target }) => {
          onChange(`res.${field.name}`)(
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
        value={value || ''}
        onChange={onChange(`res.${field.name}`)}
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
                value={parseFloat(value.min) / (field.multiplier || 1) || ''}
                onChange={({ target }) => {
                  onChange(`res.${field.name}.min`)(
                    dummyEvent(
                      `${parseFloat(target.value) * (field.multiplier || 1)}${
                        field.unit
                      }`
                    )
                  );
                  if (qos) {
                    onChange(`res.${field.name}.max`)(
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
                    onChange(`res.${field.name}.max`)(
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
                  onChange(`res.${field.name}.max`)(dummyEvent(`${value.min}`));
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

const flatM = (obj: Record<string, any>) => {
  const flatJson = {};
  for (const key in obj) {
    const parts = key.split('.');

    let temp: Record<string, any> = flatJson;

    if (parts.length === 1) {
      temp[key] = null;
    } else {
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
  }

  return flatJson;
};

type ISelectedResource = {
  label: string;
  value: string;
  resource: IMSvTemplate['resources'][number];
};

type ISelectedService = {
  label: string;
  value: string;
  service: ExtractNodeType<IProjectMSvs>;
};

interface ITemplateView {
  handleSubmit: FormEventHandler<HTMLFormElement>;
  values: Record<string, any>;
  errors: Record<string, any>;
  resources: {
    label: string;
    value: string;
    resource: ExtractNodeType<IMSvTemplate>['resources'][number];
  }[];
  services: ExtractNodeType<IProjectMSvs>[];
  isLoading: boolean;
  handleChange: (key: string) => (e: { target: { value: any } }) => void;
}

const TemplateView = ({
  handleSubmit,
  values,
  handleChange,
  errors,
  services,
  resources,
  isLoading,
}: ITemplateView) => {
  return (
    <form className="flex flex-col gap-3xl" onSubmit={handleSubmit}>
      <div className="bodyMd text-text-soft">Create your managed services.</div>

      <Select
        label="Service"
        size="lg"
        placeholder="Select service"
        value={values.selectedService?.value}
        searchable
        onChange={(val) => {
          handleChange('selectedService')(dummyEvent(val));
          handleChange('selectedResource')(dummyEvent(undefined));
        }}
        options={async () => [
          ...services.map((mt) => ({
            label: mt.displayName,
            value: parseName(mt),
            service: mt,
          })),
        ]}
        error={!!errors.selectedService}
        message={errors.selectedService}
      />
      <Select
        disabled={!values.selectedService?.value}
        label="Resource type"
        size="lg"
        placeholder="Select resource type"
        value={values.selectedResource?.value}
        searchable
        onChange={(val) => {
          handleChange('selectedResource')(dummyEvent(val));
        }}
        options={async () => [...resources]}
        error={!!values.selectedService && !!errors.selectedResource}
        message={values.selectedService ? errors.selectedResource : null}
      />
      <BottomNavigation
        primaryButton={{
          loading: isLoading,
          variant: 'primary',
          content: 'Next',
          type: 'submit',
        }}
      />
    </form>
  );
};

const FieldView = ({
  selectedResource,
  values,
  handleSubmit,
  handleChange,
  errors,
}: {
  handleChange: (key: string) => (e: { target: { value: any } }) => void;
  handleSubmit: FormEventHandler<HTMLFormElement>;
  values: Record<string, any>;
  errors: Record<string, any>;
  selectedResource: ISelectedResource | null;
}) => {
  return (
    <form
      className="flex flex-col gap-3xl"
      onSubmit={(e) => {
        if (!values.isNameError) {
          handleSubmit(e);
        } else {
          e.preventDefault();
        }
      }}
    >
      <NameIdView
        placeholder="Enter managed service name"
        label="Name"
        resType="project_managed_service"
        name={values.name}
        displayName={values.displayName}
        errors={errors.name}
        handleChange={handleChange}
        nameErrorLabel="isNameError"
      />
      {selectedResource?.resource?.fields?.map((field) => {
        const k = field.name;
        const x = k.split('.').reduce((acc, curr) => {
          if (!acc) {
            return values.res?.[curr];
          }
          return acc[curr];
        }, null);

        return (
          <RenderField
            field={field}
            key={field.name}
            onChange={handleChange}
            value={x}
            error={!!errors[k]}
            message={errors[k]}
          />
        );
      })}
      <BottomNavigation
        primaryButton={{
          variant: 'primary',
          content: 'Next',
          type: 'submit',
        }}
      />
    </form>
  );
};

const ReviewView = ({
  handleSubmit,
  values,
  selectedResource,
  selectedService,
  isLoading,
  onEdit,
}: {
  values: Record<string, any>;
  handleSubmit: FormEventHandler<HTMLFormElement>;
  selectedResource: ISelectedResource | null;
  selectedService: ISelectedService | null;
  isLoading?: boolean;
  onEdit: (step: number) => void;
}) => {
  const renderFieldView = () => {
    const fields = Object.entries(values.res).filter(
      ([k, _v]) => !['resources'].includes(k)
    );
    if (fields.length > 0) {
      return (
        <ReviewComponent
          title="Fields"
          onEdit={() => {
            onEdit(2);
          }}
        >
          <div className="flex flex-col p-xl  gap-lg rounded border border-border-default flex-1 overflow-hidden">
            {fields?.map(([key, value]) => {
              const k = key as string;
              const v = value as string;
              return (
                <div
                  key={k}
                  className="flex flex-col gap-md  [&:not(:last-child)]:pb-lg   [&:not(:last-child)]:border-b border-border-default"
                >
                  <div className="bodyMd-medium text-text-default">
                    {titleCase(k)}
                  </div>
                  <div className="bodySm text-text-soft">{v}</div>
                </div>
              );
            })}
          </div>
        </ReviewComponent>
      );
    }
    return null;
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-3xl">
      <div className="flex flex-col gap-xl">
        <ReviewComponent
          title="Basic detail"
          onEdit={() => {
            onEdit(2);
          }}
        >
          <div className="flex flex-col p-xl gap-lg rounded border border-border-default">
            <div className="flex flex-col gap-md">
              <div className="bodyMd-semibold text-text-default">
                {values.displayName}
              </div>
              <div className="bodySm text-text-soft">{values.name}</div>
            </div>
          </div>
        </ReviewComponent>
        {selectedResource && (
          <ReviewComponent
            title="Service detail"
            onEdit={() => {
              onEdit(1);
            }}
          >
            <div className="flex flex-col p-xl gap-lg rounded border border-border-default">
              <div className="flex flex-col gap-md pb-lg border-b border-border-default">
                <div className="bodyMd-semibold text-text-default">Service</div>
                <div className="bodySm text-text-soft">
                  {selectedService?.service?.metadata?.name}
                </div>
              </div>
              <div className="flex flex-col gap-md">
                <div className="bodyMd-semibold text-text-default">
                  Resource type
                </div>
                <div className="bodySm text-text-soft">
                  {selectedResource?.resource?.name}
                </div>
              </div>
            </div>
          </ReviewComponent>
        )}
        {renderFieldView()}
      </div>
      <BottomNavigation
        primaryButton={{
          variant: 'primary',
          content: 'Create',
          loading: isLoading,
          type: 'submit',
        }}
      />
    </form>
  );
};

const App = ({ services }: { services: ExtractNodeType<IProjectMSvs>[] }) => {
  const { msvtemplates } = useOutletContext<IProjectContext>();
  const navigate = useNavigate();
  const api = useConsoleApi();

  const { project, account, environment } = useParams();
  const rootUrl = `/${account}/${project}/env/${environment}/managed-resources`;

  const { currentStep, jumpStep, nextStep } = useMultiStepProgress({
    defaultStep: 1,
    totalSteps: 3,
  });

  const { values, errors, handleSubmit, handleChange, isLoading, setValues } =
    useForm({
      initialValues: {
        name: '',
        displayName: '',
        selectedService: null,
        selectedResource: null,
        res: {},
        isNameError: false,
      },
      validationSchema: Yup.object({
        name: Yup.string().test('required', 'Name is required', (v) => {
          return !(currentStep === 2 && !v);
        }),
        displayName: Yup.string().test('required', 'Name is required', (v) => {
          return !(currentStep === 2 && !v);
        }),
        selectedService: Yup.object().required('Service is required'),
        selectedResource: Yup.object({}).required('Resource type is required'),
      }),
      onSubmit: async (val) => {
        const selectedService =
          val.selectedService as unknown as ISelectedService;
        const selectedResource =
          val.selectedResource as unknown as ISelectedResource;
        const submit = async () => {
          try {
            if (!project || !environment) {
              throw new Error('Project and Environment is required!.');
            }
            if (
              !selectedService ||
              (selectedService && !selectedService.service)
            ) {
              throw new Error('Service apiversion or kind error.');
            }
            const { errors: e } = await api.createManagedResource({
              projectName: project,
              envName: environment,
              mres: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },

                spec: {
                  resourceTemplate: {
                    apiVersion: selectedResource.resource.apiVersion || '',
                    kind: selectedResource.resource.kind || '',
                    spec: {
                      ...val.res,
                    },
                    msvcRef: {
                      name: parseName(selectedService?.service),
                      namespace:
                        selectedService?.service?.spec?.targetNamespace || '',
                      apiVersion:
                        selectedService?.service?.spec?.msvcSpec.serviceTemplate
                          .apiVersion || '',
                      kind:
                        selectedService?.service?.spec?.msvcSpec.serviceTemplate
                          .kind || '',
                    },
                  },
                },
              },
            });
            if (e) {
              throw e[0];
            }
            navigate(rootUrl);
          } catch (err) {
            handleError(err);
          }
        };
        switch (currentStep) {
          case 1:
            nextStep();
            break;
          case 2:
            nextStep();
            break;
          case 3:
            await submit();
            break;
          default:
            break;
        }
      },
    });

  useEffect(() => {
    const selectedResource =
      values?.selectedResource as unknown as ISelectedResource;
    if (selectedResource?.resource?.fields) {
      setValues({
        ...values,
        res: {
          ...flatM(
            selectedResource?.resource?.fields.reduce((acc, curr) => {
              return { ...acc, [curr.name]: curr.defaultValue };
            }, {})
          ),
        },
      });
    }
  }, [values.selectedResource]);

  const resources = useMapper(
    [
      ...(getManagedTemplate({
        templates: msvtemplates || [],
        kind:
          (values.selectedService as unknown as ISelectedService)?.service?.spec
            ?.msvcSpec.serviceTemplate.kind || '',
        apiVersion:
          (values.selectedService as unknown as ISelectedService)?.service?.spec
            ?.msvcSpec.serviceTemplate.apiVersion || '',
      })?.resources || []),
    ],
    (res) => ({
      label: res.displayName,
      value: res.name,
      resource: res,
    })
  );

  return (
    <MultiStepProgressWrapper
      title="Let’s create new managed resource."
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite teams"
      backButton={{
        content: 'Back to managed resources',
        to: rootUrl,
      }}
    >
      <MultiStepProgress.Root currentStep={currentStep} jumpStep={jumpStep}>
        <MultiStepProgress.Step label="Select service" step={1}>
          <TemplateView
            isLoading={isLoading}
            services={services}
            handleChange={handleChange}
            handleSubmit={handleSubmit}
            errors={errors}
            values={values}
            resources={resources}
          />
        </MultiStepProgress.Step>
        <MultiStepProgress.Step label="Configure managed resource" step={2}>
          <FieldView
            selectedResource={values.selectedResource}
            values={values}
            errors={errors}
            handleChange={handleChange}
            handleSubmit={handleSubmit}
          />
        </MultiStepProgress.Step>
        <MultiStepProgress.Step label="Review" step={3}>
          <ReviewView
            onEdit={jumpStep}
            values={values}
            handleSubmit={handleSubmit}
            selectedService={values.selectedService}
            selectedResource={values.selectedResource}
            isLoading={isLoading}
          />
        </MultiStepProgress.Step>
      </MultiStepProgress.Root>
    </MultiStepProgressWrapper>
  );
};

const ManagedServiceLayout = () => {
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ managedServicesData }) => {
        const managedServicesList = parseNodes(managedServicesData);
        return <App services={managedServicesList} />;
      }}
    </LoadingComp>
  );
};

const NewManagedService = () => {
  return <ManagedServiceLayout />;
};

export const handle = {
  noMainLayout: true,
};

export default NewManagedService;
