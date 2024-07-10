/* eslint-disable guard-for-in */
/* eslint-disable react/destructuring-assignment */
import { useOutletContext, useParams } from '@remix-run/react';
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
import useCustomSwr from '~/lib/client/hooks/use-custom-swr';
import { toast } from '~/components/molecule/toast';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import ListV2 from '~/console/components/listV2';
import { ListItem } from '~/console/components/console-list-components';
import { CopyContentToClipboard } from '~/console/components/common-console-components';
import { LoadingPlaceHolder } from '~/console/components/loading';
import { IEnvironmentContext } from '~/console/routes/_main+/$account+/env+/$environment+/_layout';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';

type BaseType = ExtractNodeType<IManagedResources>;
type IDialog = IDialogBase<BaseType> & {
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
        resType="managed_service"
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

  const { msv } = useParams();

  const { values, errors, handleChange, handleSubmit, isLoading } = useForm({
    initialValues: isUpdate
      ? {
          name: parseName(props.data),
          displayName: props.data.displayName,
          namespace: props.data.metadata?.namespace,
          isNameError: false,
          res: {
            ...props.data.spec?.resourceTemplate.msvcRef,
          },
        }
      : {
          name: '',
          displayName: '',
          res: {},
          isNameError: false,
          namespace: '',
        },
    validationSchema: Yup.object({}),
    onSubmit: async (val) => {
      if (isUpdate) {
        try {
          if (!msv) {
            throw new Error('Managed Service is required!.');
          }
          const { errors: e } = await api.updateManagedResource({
            msvcName: msv,
            mres: {
              displayName: val.displayName,
              metadata: {
                name: val.name,
                namespace: val.namespace,
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

export const ViewSecret = ({
  show,
  setShow,
  item,
}: {
  show: boolean;
  setShow: () => void;
  item: BaseType;
}) => {
  const api = useConsoleApi();
  const { environment } = useOutletContext<IEnvironmentContext>();
  const [onClick, setOnClick] = useState(false);
  const params = useParams();
  ensureAccountClientSide(params);
  const { data, isLoading, error } = useCustomSwr(
    () =>
      onClick ? `secret _${item.syncedOutputSecretRef?.metadata?.name}` : null,
    async () => {
      if (!item.syncedOutputSecretRef?.metadata?.name) {
        toast.error('Secret not found');
        throw new Error('Secret not found');
      } else {
        return api.getSecret({
          envName: parseName(environment),

          name: item.syncedOutputSecretRef?.metadata?.name,
        });
      }
    }
  );

  const dataSecret = () => {
    if (isLoading) {
      return <LoadingPlaceHolder />;
    }
    if (error) {
      return (
        <span className="bodyMd-medium text-text-strong">
          Error while fetching secrets
        </span>
      );
    }
    if (!data?.stringData) {
      return (
        <span className="bodyMd-medium text-text-strong">No secret found</span>
      );
    }

    return (
      <ListV2.Root
        data={{
          headers: [
            {
              render: () => 'Key',
              name: 'key',
              className: 'min-w-[170px]',
            },
            {
              render: () => 'Value',
              name: 'value',
              className: 'flex-1 min-w-[345px] max-w-[345px] w-[345px]',
            },
          ],
          rows: Object.entries(data.stringData || {}).map(([key, value]) => {
            const v = value as string;
            return {
              columns: {
                key: {
                  render: () => <ListItem data={key} />,
                },
                value: {
                  render: () => (
                    <CopyContentToClipboard
                      content={v}
                      toastMessage={`${key} copied`}
                    />
                  ),
                },
              },
            };
          }),
        }}
      />
    );
  };

  useEffect(() => {
    if (error) {
      toast.error(error);
    }
  }, [error]);

  const { onNext, currentStep } = useMultiStep({
    defaultStep: 0,
    totalSteps: 2,
  });
  return (
    <Popup.Root show={show} onOpenChange={setShow}>
      <Popup.Header>
        {currentStep === 0 ? <div>Confirmation</div> : <div>Secrets</div>}
      </Popup.Header>
      <Popup.Content>
        <MultiStep.Root currentStep={currentStep}>
          <MultiStep.Step step={0}>
            <div>
              <p>{`Are you sure you want to view the secrets of '${item.syncedOutputSecretRef?.metadata?.name}'?`}</p>
            </div>
          </MultiStep.Step>
          <MultiStep.Step step={1}>{dataSecret()}</MultiStep.Step>
        </MultiStep.Root>
      </Popup.Content>
      <Popup.Footer>
        {currentStep === 0 ? (
          <Popup.Button
            content="Yes"
            onClick={() => {
              onNext();
              setOnClick(true);
            }}
          />
        ) : null}
      </Popup.Footer>
    </Popup.Root>
  );
};
