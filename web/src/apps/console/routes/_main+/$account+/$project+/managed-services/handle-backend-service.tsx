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
import { IProjectMSvs } from '~/console/server/gql/queries/project-managed-services-queries';
import { getManagedTemplate } from '~/console/utils/commons';

type IDialog = IDialogBase<ExtractNodeType<IProjectMSvs>> & {
  templates: IMSvTemplates;
};

type ISelectedService = {
  category: {
    name: string;
    displayName: string;
  };

  service: NN<IMSvTemplates>[number]['items'][number];
} | null;

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

const flatM = (obj: Record<string, any>) => {
  const flatJson = {};
  for (const key in obj) {
    const parts = key.split('.');
    let temp: Record<string, any> = flatJson;

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
  const [selectedService, setSelectedService] =
    useState<ISelectedService>(null);

  const api = useConsoleApi();
  const reload = useReload();

  const { project } = useParams();
  const [step, setStep] = useState<'choose' | 'fill'>('choose');

  const {
    values,
    errors,
    handleChange,
    handleSubmit,
    resetValues,
    isLoading,
    setValues,
  } = useForm({
    initialValues: isUpdate
      ? {
          name: '',
          displayName: '',
          res: {
            ...props.data.spec?.msvcSpec.serviceTemplate.spec,
          },
        }
      : {
          name: '',
          displayName: '',
          res: {},
        },
    validationSchema: Yup.object({}),
    onSubmit: async (val) => {
      const tempVal = { ...val };
      // @ts-ignore
      delete tempVal.name;
      // @ts-ignore
      delete tempVal.displayName;

      try {
        if (!project) {
          throw new Error('Project is required!.');
        }
        if (
          !selectedService?.service.apiVersion ||
          !selectedService.service.kind
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
                  apiVersion: selectedService.service.apiVersion,
                  kind: selectedService.service.kind,
                  spec: {
                    ...tempVal,
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
        setVisible(false);
        reload();
      } catch (err) {
        handleError(err);
      }
    },
  });

  const getService = () => {
    if (isUpdate)
      return getManagedTemplate({
        templates,
        apiVersion: props.data.spec?.msvcSpec.serviceTemplate.apiVersion || '',
        kind: props.data.spec?.msvcSpec.serviceTemplate.kind || '',
      });
    return null;
  };
  if (!isUpdate && getService()) {
    return null;
  }
  return (
    <Popup.Form
      onSubmit={(e) => {
        if (step === 'choose') {
          setStep('fill');
          e.preventDefault();
        } else handleSubmit(e);
      }}
    >
      <Popup.Content className="!min-h-[500px] !max-h-[500px]">
        <Fill
          {...{
            templates,
            selectedService: getService(),
            values,
            errors,
            handleChange,
          }}
        />
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
        {isUpdate ? 'Edit managed service' : 'Add managed service'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleBackendService;
