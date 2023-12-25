import { useOutletContext } from '@remix-run/react';
import { useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import Select from '~/components/atoms/select';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { IdSelector } from '~/console/components/id-selector';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import RenderDynamicField from '~/console/components/render-dynamic-field';
import { IDialog } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IManagedServiceTemplate } from '~/console/server/gql/queries/managed-service-queries';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { IManagedServiceContext } from '../_index';

const HandleBackendResources = ({
  show,
  setShow,
  template,
}: IDialog & { template: IManagedServiceTemplate }) => {
  const [isIDLoading, setIsIDLoading] = useState(false);
  const { currentStep, onNext, onPrevious } = useMultiStep({
    defaultStep: 1,
    totalSteps: 2,
  });

  const [selectedType, setSelectedType] = useState<{
    label: string;
    value: string;
    resource: IManagedServiceTemplate['resources'][number];
  }>();

  const api = useConsoleApi();
  const { backendService, workspace } =
    useOutletContext<IManagedServiceContext>();

  const {
    values: valuesFirst,
    errors: errorsFirst,
    handleChange: handleChangeFirst,
    handleSubmit: handleSubmitFirst,
  } = useForm({
    initialValues: {
      name: '',
      displayName: '',
      type: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required(),
      displayName: Yup.string().required(),
      type: Yup.string().required(),
    }),
    onSubmit: () => {
      onNext();
    },
  });

  const {
    values: valuesSecond,
    errors: errorsSecond,
    handleChange: handleChangeSecond,
    handleSubmit: handleSubmitSecond,
  } = useForm({
    initialValues: {
      ...(selectedType?.resource.fields.reduce((acc, field) => {
        return {
          ...acc,
          [field.name]: field.defaultValue,
        };
      }, {}) || {}),
    },
    validationSchema: Yup.object({
      ...selectedType?.resource.fields.reduce((acc, curr) => {
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
      // try {
      //   const { errors: e } = await api.createManagedResource({
      //     mres: {
      //       displayName: valuesFirst.displayName,
      //       metadata: {
      //         name: valuesFirst.name,
      //         namespace: parseTargetNs(workspace),
      //       },
      //       spec: {
      //         mresKind: {
      //           kind: selectedType?.resource.kind || '',
      //         },
      //         msvcRef: {
      //           apiVersion: template.apiVersion || '',
      //           name: parseName(backendService),
      //           kind: template.kind!,
      //         },
      //         inputs: {
      //           ...val,
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
      show={show as any}
      onOpenChange={(e) => {
        setShow(e);
      }}
    >
      <Popup.Header>Create new resource</Popup.Header>
      <form
        onSubmit={currentStep === 1 ? handleSubmitFirst : handleSubmitSecond}
      >
        <Popup.Content>
          <MultiStep.Root currentStep={currentStep}>
            <MultiStep.Step step={1}>
              <div className="flex flex-col">
                <div className="pb-3xl">
                  <TextInput
                    value={valuesFirst.displayName}
                    label="Name"
                    onChange={handleChangeFirst('displayName')}
                    error={!!errorsFirst.displayName}
                    message={errorsFirst.displayName}
                  />
                  <IdSelector
                    resType="managed_resource"
                    onLoad={(loading) => setIsIDLoading(loading)}
                    onChange={(v) => {
                      handleChangeFirst('name')(dummyEvent(v));
                    }}
                    name={valuesFirst.displayName}
                    className="pt-2xl"
                  />
                </div>

                <Select
                  label="Type"
                  placeholder="--- Select type ---"
                  value={selectedType}
                  options={async () =>
                    template.resources.map((tr) => ({
                      label: tr.displayName,
                      value: tr.name,
                      resource: tr,
                    }))
                  }
                  onChange={(value) => {
                    setSelectedType(value);
                    handleChangeFirst('type')(dummyEvent(value.value));
                  }}
                  error={!!errorsFirst.type}
                  message={errorsFirst.type}
                />
              </div>
            </MultiStep.Step>
            <MultiStep.Step step={2}>
              <div className="flex flex-col">
                {selectedType?.resource.fields.map((field) => {
                  const { name } = field;
                  return (
                    <RenderDynamicField
                      key={name}
                      field={field}
                      value={(valuesSecond as Record<string, any>)[name]}
                      error={!!errorsSecond[name]}
                      message={errorsSecond[name]}
                      onChange={handleChangeSecond(name)}
                    />
                  );
                })}
              </div>
            </MultiStep.Step>
          </MultiStep.Root>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button
            type="button"
            variant="basic"
            content="Back"
            onClick={onPrevious}
          />
          <Popup.Button
            type="submit"
            content="Create"
            variant="primary"
            disabled={isIDLoading}
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

export default HandleBackendResources;
