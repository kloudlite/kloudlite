/* eslint-disable guard-for-in */
/* eslint-disable react/destructuring-assignment */
import { useParams } from '@remix-run/react';
import { useEffect, useRef, useState } from 'react';
import { NumberInput, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { IDialogBase } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import { useReload } from '~/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/lib/client/hooks/use-form';
import Yup from '~/lib/server/helpers/yup';
import { NN } from '~/lib/types/common';
import { handleError } from '~/lib/utils/common';
import {
  IMSvTemplate,
  IMSvTemplates,
} from '~/console/server/gql/queries/managed-templates-queries';
import { Switch } from '~/components/atoms/switch';
import { getManagedTemplate } from '~/console/utils/commons';
import { NameIdView } from '~/console/components/name-id-view';
import { IManagedResources } from '~/console/server/gql/queries/managed-resources-queries';

type IDialog = IDialogBase<ExtractNodeType<IManagedResources>> & {
  templates: IMSvTemplates;
};

type ISelectedService = {
  category: {
    name: string;
    displayName: string;
  };

  service?: NN<IMSvTemplates>[number]['items'][number];
} | null;

const RenderField = ({
  field,
  value,
  onChange,
  errors,
  fieldKey,
}: {
  field: NN<NN<ISelectedService>['service']>['fields'][number];
  onChange: (e: string) => (e: { target: { value: any } }) => void;
  value: any;
  errors: Record<string, any>;
  fieldKey: string;
}) => {
  const [qos, setQos] = useState(false);

  useEffect(() => {
    if (field.inputType === 'Resource' && value.max === value.min) {
      setQos(true);
    }
  }, []);

  if (field.inputType === 'Number') {
    return (
      <NumberInput
        error={!!errors[fieldKey]}
        message={errors[fieldKey]}
        label={`${field.label}${field.required ? ' *' : ''}`}
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
        error={!!errors[fieldKey]}
        message={errors[fieldKey]}
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
                error={!!errors[`${fieldKey}.min`]}
                message={errors[`${fieldKey}.min`]}
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
                  error={!!errors[`${fieldKey}.max`]}
                  message={errors[`${fieldKey}.max`]}
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

type ISelectedResource = IMSvTemplate['resources'][number];

const Fill = ({
  selectedResource,
  values,
  handleChange,
  errors,
}: {
  selectedResource: ISelectedResource | null | undefined;
  values: { [key: string]: any };
  handleChange: (key: string) => (e: { target: { value: any } }) => void;
  errors: Record<string, any>;
}) => {
  const nameRef = useRef<HTMLInputElement>(null);
  useEffect(() => {
    nameRef.current?.focus();
  }, [nameRef.current]);
  return (
    <div className="flex flex-col gap-3xl">
      <NameIdView
        placeholder="Enter managed service name"
        label="Name"
        resType="project_managed_service"
        name={values.name}
        displayName={values.displayName}
        errors={errors.name}
        handleChange={handleChange}
        nameErrorLabel="isNameError"
        isUpdate
      />
      {selectedResource?.fields?.map((field) => {
        const k = field.name;
        const x = k.split('.').reduce((acc, curr) => {
          if (!acc) {
            return values.res?.[curr];
          }
          return acc[curr];
        }, null);

        return (
          <RenderField
            errors={errors}
            value={x}
            onChange={handleChange}
            key={k}
            field={field}
            fieldKey={k}
          />
        );
      })}
    </div>
  );
};

const Root = (props: IDialog) => {
  const { isUpdate, setVisible, templates } = props;

  const api = useConsoleApi();
  const reload = useReload();

  const { project, environment } = useParams();

  const { values, errors, handleChange, handleSubmit, isLoading } = useForm({
    initialValues: isUpdate
      ? {
          name: parseName(props.data),
          displayName: props.data.displayName,
          isNameError: false,
          res: {
            ...props.data.spec?.resourceTemplate.spec,
          },
        }
      : {
          name: '',
          displayName: '',
          res: {},
          isNameError: false,
        },
    validationSchema: Yup.object({}),
    onSubmit: async (val) => {
      if (isUpdate) {
        try {
          if (!project || !environment) {
            throw new Error('Project and environment is required!.');
          }
          const { errors: e } = await api.updateManagedResource({
            projectName: project,
            envName: environment,
            mres: {
              displayName: val.displayName,
              metadata: {
                name: val.name,
              },

              spec: {
                resourceTemplate: {
                  ...props.data.spec.resourceTemplate,
                  spec: {
                    ...val.res,
                  },
                },
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
      }
    },
  });

  const getResources = () => {
    if (isUpdate)
      return getManagedTemplate({
        templates,
        apiVersion: props.data.spec?.resourceTemplate.msvcRef.apiVersion || '',
        kind: props.data.spec?.resourceTemplate.msvcRef.kind || '',
      })?.resources.find(
        (rs) =>
          rs.apiVersion === props.data.spec.resourceTemplate.apiVersion &&
          rs.kind === props.data.spec.resourceTemplate.kind
      );
    return undefined;
  };

  if (!isUpdate) {
    return null;
  }
  return (
    <Popup.Form
      onSubmit={(e) => {
        if (!values.isNameError) {
          handleSubmit(e);
        } else {
          e.preventDefault();
        }
      }}
    >
      <Popup.Content>
        <Fill
          {...{
            templates,
            selectedResource: getResources(),
            values,
            errors,
            handleChange,
          }}
        />
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button type="button" variant="basic" content="Cancel" closable />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content="Update"
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleManagedResources = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;
  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>
        {isUpdate ? 'Edit managed resource' : 'Add managed resource'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleManagedResources;
