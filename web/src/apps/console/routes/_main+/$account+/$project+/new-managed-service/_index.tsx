/* eslint-disable guard-for-in */
import { ArrowRight } from '@jengaicons/react';
import { useNavigate, useOutletContext, useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import Select from '~/components/atoms/select';
import { NameIdView } from '~/console/components/name-id-view';
import ProgressWrapper from '~/console/components/progress-wrapper';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { FormEventHandler, useEffect, useRef, useState } from 'react';
import {
  IMSvTemplate,
  IMSvTemplates,
} from '~/console/server/gql/queries/managed-templates-queries';
import { Switch } from '~/components/atoms/switch';
import { NumberInput, TextInput } from '~/components/atoms/input';
import { handleError } from '~/root/lib/utils/common';
import { titleCase } from '~/components/utils';
import { IProjectContext } from '../_layout';
import { ReviewComponent } from '../$environment+/new-app/app-review';

const valueRender = ({ label, icon }: { label: string; icon: string }) => {
  return (
    <div className="flex flex-row gap-lg items-center">
      <span>
        <img alt={label} src={icon} className="w-2xl h-w-2xl" />
      </span>
      <div>{label}</div>
    </div>
  );
};

type steps = 'Select template' | 'Configure managed service' | 'Review';

const hasTemplate = (res: any) =>
  (!!res && res?.template?.fields.length > 0) || !res;

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
                value={parseFloat(value.min) / (field.multiplier || 1)}
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

const flatMapValidations = (obj: Record<string, any>) => {
  console.log('validations', obj);

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

type ISelectedTemplate = {
  category: string;
  categoryDisplayName: string;
  template: IMSvTemplate;
};

const TemplateView = ({
  handleSubmit,
  values,
  handleChange,
  errors,
  templates,
  isLoading,
}: {
  handleSubmit: FormEventHandler<HTMLFormElement>;
  values: Record<string, any>;
  errors: Record<string, any>;
  templates: IMSvTemplates;
  isLoading: boolean;
  handleChange: (key: string) => (e: { target: { value: any } }) => void;
}) => {
  const nameRef = useRef<HTMLInputElement>(null);
  useEffect(() => {
    nameRef.current?.focus();
  }, [nameRef.current]);
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
        ref={nameRef}
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
        label="Template"
        size="lg"
        placeholder="Select templates"
        value={
          values.selectedTemplate
            ? { label: '', value: values.selectedTemplate?.template.name || '' }
            : undefined
        }
        valueRender={valueRender}
        searchable
        error={!!errors.selectedTemplate}
        message={errors.selectedTemplate}
        onChange={({ item }) => {
          handleChange('selectedTemplate')(dummyEvent(item));
        }}
        options={async () =>
          templates.map((mt) => ({
            label: mt.displayName,
            options: mt.items.map((mti) => ({
              label: mti.displayName,
              value: mti.name,
              icon: mti.logoUrl,
              item: {
                categoryDisplayName: mt.displayName,
                category: mt.category,
                template: mti,
              },
              render: () => (
                <div className="flex flex-row items-center gap-xl">
                  <span>
                    <img
                      alt={mti.displayName}
                      src={mti.logoUrl}
                      className="w-2xl h-w-2xl"
                    />
                  </span>
                  <div>{mti.displayName}</div>
                </div>
              ),
            })),
          }))
        }
      />

      <div className="flex flex-row justify-start">
        <Button
          loading={isLoading}
          variant="primary"
          content={hasTemplate(values.selectedTemplate) ? 'Next' : 'Create'}
          suffix={<ArrowRight />}
          type="submit"
        />
      </div>
    </form>
  );
};

const FieldView = ({
  selectedTemplate,
  values,
  handleSubmit,
  handleChange,
  errors,
}: {
  handleChange: (key: string) => (e: { target: { value: any } }) => void;
  handleSubmit: FormEventHandler<HTMLFormElement>;
  values: Record<string, any>;
  errors: Record<string, any>;
  selectedTemplate: ISelectedTemplate | null;
}) => {
  return (
    <form className="flex flex-col gap-3xl py-3xl" onSubmit={handleSubmit}>
      {selectedTemplate?.template.fields?.map((field) => {
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
  isLoading,
}: {
  values: Record<string, any>;
  handleSubmit: FormEventHandler<HTMLFormElement>;
  isLoading?: boolean;
}) => {
  const renderFieldView = () => {
    const fields = Object.entries(values.res).filter(
      ([k, _v]) => !['resources'].includes(k)
    );
    if (fields.length > 0) {
      return (
        <ReviewComponent title="Fields" onEdit={() => {}}>
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
              {values?.selectedTemplate?.categoryDisplayName}
            </div>
            <div className="bodySm text-text-soft">
              {values?.selectedTemplate?.template?.displayName}
            </div>
          </div>
        </ReviewComponent>
        {renderFieldView()}
        {values?.res?.resources && (
          <ReviewComponent title="Fields" onEdit={() => {}}>
            <div className="flex flex-col p-xl  gap-lg rounded border border-border-default flex-1 overflow-hidden">
              {Object.entries(values?.res?.resources).map(([key, value]) => {
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
          </ReviewComponent>
        )}
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

const ManagedServiceLayout = () => {
  const { msvtemplates } = useOutletContext<IProjectContext>();
  const [activeState, setActiveState] = useState<steps>('Select template');
  const navigate = useNavigate();
  const api = useConsoleApi();

  const { project, account } = useParams();

  const { values, errors, handleSubmit, handleChange, isLoading, setValues } =
    useForm({
      initialValues: {
        name: '',
        displayName: '',
        res: {},
        selectedTemplate: null,
        isNameError: false,
      },
      validationSchema: Yup.object({
        name: Yup.string().required(),
        displayName: Yup.string().required(),
        selectedTemplate: Yup.object({}).required('Template is required.'),
      }),
      onSubmit: async (val) => {
        const selectedTemplate =
          val.selectedTemplate as unknown as ISelectedTemplate;

        const submit = async () => {
          try {
            if (!project) {
              throw new Error('Project is required!.');
            }
            if (
              !selectedTemplate?.template?.apiVersion ||
              !selectedTemplate?.template?.kind
            ) {
              throw new Error('Service apiversion or kind error.');
            }
            const { errors: e } = await api.createProjectMSv({
              projectName: project,
              pmsvc: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },

                spec: {
                  msvcSpec: {
                    serviceTemplate: {
                      apiVersion: selectedTemplate.template.apiVersion,
                      kind: selectedTemplate.template.kind,
                      spec: {
                        ...val.res,
                      },
                    },
                  },
                  targetNamespace: '',
                },
              },
            });
            if (e) {
              throw e[0];
            }
            navigate(`/${account}/${project}/managed-services`);
          } catch (err) {
            handleError(err);
          }
        };

        switch (activeState) {
          case 'Select template':
            if (!hasTemplate(val.selectedTemplate)) {
              await submit();
              break;
            }
            setActiveState('Configure managed service');
            break;
          case 'Configure managed service':
            if (!hasTemplate(val.selectedTemplate)) {
              await submit();
              break;
            }
            setActiveState('Review');
            break;
          case 'Review':
            submit();
            break;
          default:
            break;
        }
      },
    });

  useEffect(() => {
    const selectedTemplate =
      values.selectedTemplate as unknown as ISelectedTemplate;
    if (selectedTemplate?.template?.fields) {
      setValues({
        ...values,
        res: {
          ...flatM(
            selectedTemplate?.template?.fields.reduce((acc, curr) => {
              return { ...acc, [curr.name]: curr.defaultValue };
            }, {})
          ),
        },
      });
    }
  }, [values.selectedTemplate]);

  useEffect(() => {
    console.log(
      'validation: ',
      flatMapValidations(
        values.selectedTemplate?.template?.fields.reduce((acc, curr) => {
          return {
            ...acc,
            [curr.name]: curr,
          };
        }, {})
      )
    );
  }, [values.res]);

  // (() => {
  //   let returnYup: any = Yup;
  //   switch (curr.inputType) {
  //     case 'Number':
  //       returnYup = returnYup.number();
  //       if (curr.min) returnYup = returnYup.min(curr.min);
  //       if (curr.max) returnYup = returnYup.max(curr.max);
  //       break;
  //     case 'String':
  //       returnYup = returnYup.string();
  //       break;
  //     default:
  //       returnYup = returnYup.string();
  //   }

  //   if (curr.required) {
  //     returnYup = returnYup.required();
  //   }

  //   return returnYup;
  // })()
  const isActive = (step: steps) => step === activeState;

  const getItems = () => {
    return [
      {
        label: 'Select template',
        active: isActive('Select template'),
        completed: false,
        children: isActive('Select template') ? (
          <TemplateView
            isLoading={isLoading}
            templates={msvtemplates}
            handleChange={handleChange}
            handleSubmit={handleSubmit}
            errors={errors}
            values={values}
          />
        ) : null,
      },
      ...(hasTemplate(values.selectedTemplate)
        ? [
            {
              label: 'Configure managed service',
              active: isActive('Configure managed service'),
              completed: false,
              children: isActive('Configure managed service') ? (
                <FieldView
                  selectedTemplate={values.selectedTemplate}
                  values={values.res}
                  errors={errors}
                  handleChange={handleChange}
                  handleSubmit={handleSubmit}
                />
              ) : null,
            },
          ]
        : []),
      ...(hasTemplate(values.selectedTemplate)
        ? [
            {
              label: 'Review',
              active: isActive('Review'),
              completed: false,
              children: isActive('Review') ? (
                <ReviewView
                  values={values}
                  handleSubmit={handleSubmit}
                  isLoading
                />
              ) : null,
            },
          ]
        : []),
    ];
  };

  return (
    <ProgressWrapper
      title="Let’s create new managed service."
      subTitle="Simplify Collaboration and Enhance Productivity with Kloudlite teams"
      progressItems={{
        items: getItems(),
      }}
      onClick={({ label }) => {}}
    />
  );
};

const NewManagedService = () => {
  return <ManagedServiceLayout />;
};

export const handle = {
  noMainLayout: true,
};

export default NewManagedService;
