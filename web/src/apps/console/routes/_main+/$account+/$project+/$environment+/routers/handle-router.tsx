/* eslint-disable react/destructuring-assignment */
import { useParams } from '@remix-run/react';
import { useEffect, useRef, useState } from 'react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import CommonPopupHandle from '~/console/components/common-popup-handle';
import { IDialogBase } from '~/console/components/types.d';
import { IRouters } from '~/console/server/gql/queries/router-queries';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { NameIdView } from '~/console/components/name-id-view';
import Select from '~/components/atoms/select';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { IDomains } from '~/console/server/gql/queries/domain-queries';
import { useMapper } from '~/components/utils';

type IDialog = IDialogBase<ExtractNodeType<IRouters>>;

const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();

  const { project: projectName, environment: envName } = useParams();
  const [selectedDomains, setSelectedDomains] = useState<
    { label: string; value: string; domain: ExtractNodeType<IDomains> }[]
  >([]);

  const {
    data,
    isLoading: domainLoading,
    error: domainLoadingError,
  } = useCustomSwr('/domains', async () => {
    return api.listDomains({});
  });

  const { values, errors, handleSubmit, handleChange, isLoading, resetValues } =
    useForm({
      initialValues: isUpdate
        ? {
            name: parseName(props.data),
            displayName: props.data.displayName,
            domains: [],
            isNameError: false,
          }
        : {
            name: '',
            displayName: '',
            domains: [],
            isNameError: false,
          },
      validationSchema: Yup.object({
        displayName: Yup.string().required(),
        name: Yup.string().required(),
        domains: Yup.array().test('required', 'domain is required', (val) => {
          return val && val?.length > 0;
        }),
        // .test('is-valid', 'invalid domain names', (val) => {
        //   console.log('vals', val);

        //   return val?.every((v) => v.endsWith('.com'));
        // }),
      }),

      onSubmit: async (val) => {
        if (!projectName || !envName || selectedDomains?.length === 0) {
          throw new Error('Project, Environment and Domain is required!.');
        }
        try {
          if (!isUpdate) {
            const { errors: e } = await api.createRouter({
              envName,
              projectName,
              router: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                spec: {
                  domains: selectedDomains.map((sd) => sd.value),
                  https: {
                    enabled: true,
                  },
                },
              },
            });
            if (e) {
              throw e[0];
            }
            toast.success('Router created successfully');
          } else {
            const { errors: e } = await api.updateRouter({
              envName,
              projectName,
              router: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                spec: {
                  domains: selectedDomains.map((sd) => sd.value),
                  https: {
                    enabled: true,
                  },
                },
              },
            });
            if (e) {
              throw e[0];
            }
            toast.success('Router updated successfully');
          }
          reloadPage();
          setVisible(false);
          resetValues();
        } catch (err) {
          handleError(err);
        }
      },
    });

  const domains = useMapper(parseNodes(data), (val) => ({
    label: val.displayName,
    value: val.domainName,
    domain: val,
    render: () => val.displayName,
  }));

  useEffect(() => {
    if (isUpdate) {
      setSelectedDomains(
        domains.filter((d) => props.data.spec.domains.includes(d.value))
      );
    }
  }, [isUpdate]);

  const nameIDRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    nameIDRef.current?.focus();
  }, [nameIDRef]);

  useEffect(() => {
    console.log(errors);
  }, [errors]);
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
      <Popup.Content className="flex flex-col justify-start gap-3xl">
        <NameIdView
          ref={nameIDRef}
          resType="router"
          label="Name"
          displayName={values.displayName}
          name={values.name}
          errors={errors.name}
          handleChange={handleChange}
          nameErrorLabel="isNameError"
          isUpdate={isUpdate}
        />
        <Select
          creatable
          size="lg"
          label="Domains"
          multiple
          value={selectedDomains}
          options={async () => [...domains]}
          onChange={(val) => {
            setSelectedDomains(val);
            handleChange('domains')(dummyEvent([...val.map((v) => v.value)]));
          }}
          error={!!errors.domains || !!domainLoadingError}
          message={
            errors.domains ||
            (domainLoadingError ? 'Error fetching domains.' : '')
          }
          loading={domainLoading}
          disableWhileLoading
        />
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button content="Cancel" variant="basic" closable />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content={!isUpdate ? 'Add' : 'Update'}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleRouter = (props: IDialog) => {
  return (
    <CommonPopupHandle
      {...props}
      createTitle="Create router"
      updateTitle="Update router"
      root={Root}
    />
  );
};
export default HandleRouter;
