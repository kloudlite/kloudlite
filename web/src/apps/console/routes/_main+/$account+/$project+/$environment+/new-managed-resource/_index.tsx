/* eslint-disable guard-for-in */

import { ArrowRight } from '@jengaicons/react';
import {
  useLoaderData,
  useNavigate,
  useOutletContext,
  useParams,
} from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import Select from '~/components/atoms/select';
import { NameIdView } from '~/console/components/name-id-view';
import ProgressWrapper from '~/console/components/progress-wrapper';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { FormEventHandler, useEffect, useState } from 'react';
import { IMSvTemplate } from '~/console/server/gql/queries/managed-templates-queries';
import { Switch } from '~/components/atoms/switch';
import { NumberInput, TextInput } from '~/components/atoms/input';
import { handleError } from '~/root/lib/utils/common';
import { useMapper } from '~/components/utils';
import { IRemixCtx } from '~/root/lib/types/common';
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
import { ReviewComponent } from '../new-app/app-review';
import { IProjectContext } from '../../_layout';

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

type steps = 'Select service' | 'Configure resource' | 'Review';

const hasResource = (res: any) =>
  (!!res && res?.resource?.fields.length > 0) || !res;

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
  console.log('obj', obj);

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
    <form
      className="flex flex-col gap-3xl py-3xl"
      onSubmit={(e) => {
        if (!values.isNameError) {
          handleSubmit(e);
        } else {
          e.preventDefault();
        }
      }}
    >
      <div className="bodyMd text-text-soft">Create your managed services.</div>

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
      <Select
        label="Service"
        size="lg"
        placeholder="Select service"
        value={{ label: '', value: values.selectedService?.value || '' }}
        searchable
        onChange={(val) => {
          handleChange('selectedService')(dummyEvent(val));
          handleChange('selectedResource')(dummyEvent(undefined));
          console.log(values);
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
        value={
          values.selectedResource?.value
            ? { label: '', value: values.selectedResource?.value || '' }
            : undefined
        }
        searchable
        onChange={(val) => {
          handleChange('selectedResource')(dummyEvent(val));
        }}
        options={async () => [...resources]}
        error={!!values.selectedService && !!errors.selectedResource}
        message={values.selectedService ? errors.selectedResource : null}
      />

      <div className="flex flex-row justify-start">
        <Button
          loading={isLoading}
          variant="primary"
          content={hasResource(values.selectedResource) ? 'Next' : 'Create'}
          suffix={<ArrowRight />}
          type="submit"
        />
      </div>
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
    <form className="flex flex-col gap-3xl py-3xl" onSubmit={handleSubmit}>
      {selectedResource?.resource?.fields?.map((field) => {
        const k = field.name;
        const x = k.split('.').reduce((acc, curr) => {
          if (!acc) {
            return values[curr];
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
      <div className="flex flex-row justify-start">
        <Button
          variant="primary"
          content="Next"
          suffix={<ArrowRight />}
          type="submit"
        />
      </div>
    </form>
  );
};

const ReviewView = ({
  handleSubmit,
  values,
  selectedResource,
  selectedService,
  isLoading,
}: {
  values: Record<string, any>;
  handleSubmit: FormEventHandler<HTMLFormElement>;
  selectedResource: ISelectedResource | null;
  selectedService: ISelectedService | null;
  isLoading?: boolean;
}) => {
  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-3xl py-3xl">
      <div className="flex flex-col gap-xl">
        <ReviewComponent title="Basic detail" onEdit={() => {}}>
          <div className="flex flex-col p-xl gap-lg rounded border border-border-default">
            <div className="flex flex-col gap-md">
              <div className="bodyMd-semibold text-text-default">
                {values.displayName}
              </div>
              <div className="bodySm text-text-soft">{values.name}</div>
            </div>
          </div>
        </ReviewComponent>
        <ReviewComponent title="Service detail" onEdit={() => {}}>
          <div className="flex flex-col p-xl gap-md rounded border border-border-default">
            <div className="bodyMd-semibold text-text-default">
              {selectedResource?.resource?.displayName}
            </div>
            <div className="bodySm text-text-soft">
              {selectedResource?.resource?.name}
            </div>
          </div>
        </ReviewComponent>
        {/* <ReviewComponent title="Fields" onEdit={() => {}}>
          <div className="flex flex-col p-xl  gap-lg rounded border border-border-default flex-1 overflow-hidden">
            {Object.entries(values?.resources).map(([key, value]) => {
              return (
                <div
                  key={key}
                  className="flex flex-col gap-md  [&:not(:last-child)]:pb-lg   [&:not(:last-child)]:border-b border-border-default"
                >
                  <div className="bodyMd-medium text-text-default">
                    {titleCase(key)}
                  </div>
                  <div className="bodySm text-text-soft">
                    {Object.entries(value || {}).map(([pKey, pValue]) => (
                      <div key={pKey}>
                        {pKey}
                        {' : '}
                        {pValue}
                      </div>
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        </ReviewComponent> */}
      </div>

      <div className="flex flex-row justify-start">
        <Button
          variant="primary"
          content="Create"
          loading={isLoading}
          suffix={<ArrowRight />}
          type="submit"
        />
      </div>
    </form>
  );
};

const App = ({ services }: { services: ExtractNodeType<IProjectMSvs>[] }) => {
  const { msvtemplates } = useOutletContext<IProjectContext>();
  const [activeState, setActiveState] = useState<steps>('Select service');
  const navigate = useNavigate();
  const api = useConsoleApi();

  const { project, account, environment } = useParams();

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
        name: Yup.string().required(),
        displayName: Yup.string().required(),
        selectedService: Yup.object().required(),
        selectedResource: Yup.object({}).required(),
      }),
      onSubmit: async (val) => {
        const selectedService =
          val.selectedService as unknown as ISelectedService;
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
                    spec: {
                      ...val.res,
                    },
                    msvcRef: {
                      name: parseName(selectedService?.service),
                      namespace:
                        selectedService?.service?.metadata?.namespace || '',
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
            navigate(`/${account}/${project}/${environment}/managed-resources`);
          } catch (err) {
            handleError(err);
          }
        };
        switch (activeState) {
          case 'Select service':
            if (!hasResource(val.selectedResource)) {
              await submit();
              break;
            }
            setActiveState('Configure resource');
            break;
          case 'Configure resource':
            setActiveState('Review');
            break;
          case 'Review':
            await submit();
            break;
          default:
            break;
        }
      },
    });

  useEffect(() => {
    const selectedResource =
      values.selectedResource as unknown as ISelectedResource;
    if (selectedResource.resource.fields) {
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

  const isActive = (step: steps) => step === activeState;

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

  const items = useMapper(
    [
      {
        label: 'Select service',
        active: isActive('Select service'),
        completed: false,
        children: isActive('Select service') ? (
          <TemplateView
            isLoading={isLoading}
            services={services}
            handleChange={handleChange}
            handleSubmit={handleSubmit}
            errors={errors}
            values={values}
            resources={resources}
          />
        ) : null,
      },
      ...(hasResource(values.selectedResource)
        ? [
            {
              label: 'Configure resource',
              active: isActive('Configure resource'),
              completed: false,
              children: isActive('Configure resource') ? (
                <FieldView
                  selectedResource={values.selectedResource}
                  values={values.res}
                  errors={errors}
                  handleChange={handleChange}
                  handleSubmit={handleSubmit}
                />
              ) : null,
            },
          ]
        : []),
      ...[
        ...(hasResource(values.selectedResource)
          ? [
              {
                label: 'Review',
                active: isActive('Review'),
                completed: false,
                children: isActive('Review') ? (
                  <ReviewView
                    values={values}
                    handleSubmit={handleSubmit}
                    selectedService={values.selectedService}
                    selectedResource={values.selectedResource}
                  />
                ) : null,
              },
            ]
          : []),
      ],
    ],
    (val) => val
  );
  return (
    <ProgressWrapper
      title="Let’s create new managed resource."
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite teams"
      progressItems={{
        items,
      }}
      // onClick={({ label }) => {}}
    />
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
