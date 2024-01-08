import { useParams } from '@remix-run/react';
import { useEffect, useState } from 'react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
  parseTargetNs,
} from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useDataFromMatches } from '~/root/lib/client/hooks/use-custom-matches';
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
  const [domainError, setDomainError] = useState<string | null>(null);

  const {
    data,
    isLoading: domainLoading,
    error: domainLoadingError,
  } = useCustomSwr('/domains', async () => {
    return api.listDomains({});
  });

  const {
    values,
    errors,
    handleSubmit,
    handleChange,
    isLoading,
    resetValues,
    setValues,
  } = useForm({
    initialValues: {
      name: '',
      displayName: '',
    },
    validationSchema: Yup.object({
      displayName: Yup.string().required(),
      name: Yup.string().required(),
    }),

    onSubmit: async (val) => {
      if (selectedDomains.length === 0) {
        setDomainError('Atleast one domain is required!.');
      } else {
        setDomainError(null);
      }
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
              },
            },
          });
          if (e) {
            throw e[0];
          }
          toast.success('Router created successfully');
        } else {
          //
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

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content className="flex flex-col gap-3xl">
        <NameIdView
          resType="router"
          label="Name"
          displayName={values.displayName}
          name={values.name}
          errors={errors.values}
          onChange={({ name, id }) => {
            handleChange('displayName')(dummyEvent(name));
            handleChange('name')(dummyEvent(id));
          }}
        />
        <Select
          size="lg"
          label="Domains"
          multiple
          value={selectedDomains}
          options={async () => [...domains]}
          onChange={(val) => {
            setSelectedDomains(val);
          }}
          error={!!domainError || !!domainLoadingError}
          message={domainLoadingError ? 'Error fetching domains.' : domainError}
          loading={domainLoading}
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
      createTitle="Create build"
      updateTitle="Update build"
      root={Root}
    />
  );
};
export default HandleRouter;
