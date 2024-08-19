/* eslint-disable react/destructuring-assignment */
import { toast } from 'react-toastify';
import Popup from '~/components/molecule/popup';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import { IDialogBase } from '~/console/components/types.d';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import { NameIdView } from '~/console/components/name-id-view';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IByocClusters } from '~/console/server/gql/queries/byok-cluster-queries';
import { Checkbox } from '~/components/atoms/checkbox';
import Banner from '~/components/molecule/banner';

type IDialog = IDialogBase<ExtractNodeType<IByocClusters>>;

const Root = (props: IDialog) => {
  const { setVisible, isUpdate } = props;

  const api = useConsoleApi();
  const reloadPage = useReload();

  const { values, errors, handleChange, handleSubmit, resetValues, isLoading } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: props.data.displayName,
            name: parseName(props.data),
            visibilityMode: false,
            isNameError: false,
          }
        : {
            name: '',
            displayName: '',
            visibilityMode: false,
            isNameError: false,
          },
      validationSchema: Yup.object({
        name: Yup.string().required('id is required'),
        displayName: Yup.string().required('name is required'),
      }),
      onSubmit: async (val) => {
        try {
          if (!isUpdate) {
            const { errors: e } = await api.createBYOKCluster({
              cluster: {
                displayName: val.displayName,
                metadata: {
                  name: val.name,
                },
                visibility: {
                  mode: val.visibilityMode ? 'private' : 'public',
                },
              },
            });
            if (e) {
              throw e[0];
            }
          } else if (isUpdate) {
            const { errors: e } = await api.updateByokCluster({
              clusterName: val.name,
              displayName: val.displayName,
            });
            if (e) {
              throw e[0];
            }
          }
          reloadPage();
          resetValues();
          toast.success(
            `cluster ${isUpdate ? 'updated' : 'created'} successfully`
          );
          setVisible(false);
        } catch (err) {
          handleError(err);
        }
      },
    });

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
        <div className="flex flex-col gap-2xl">
          <NameIdView
            resType="cluster"
            displayName={values.displayName}
            name={values.name}
            label="Cluster name"
            placeholder="Enter cluster name"
            errors={errors.name}
            handleChange={handleChange}
            nameErrorLabel="isNameError"
            isUpdate={isUpdate}
          />
          {!isUpdate && (
            <>
              <Checkbox
                label="Private Cluster"
                checked={values.visibilityMode}
                onChange={(val) => {
                  handleChange('visibilityMode')(dummyEvent(val));
                }}
              />
              <Banner
                type="info"
                body={
                  <div className="flex flex-col">
                    <span className="bodyMd-medium">
                      Private clusters are those who are hosted behind a NAT.
                    </span>
                    <span className="bodyMd">
                      Ex: Cluster running on your local machine
                    </span>
                  </div>
                }
              />
            </>
          )}
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content={isUpdate ? 'Update' : 'Create'}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleByokCluster = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>
        {isUpdate ? 'Edit Cluster' : 'Attach Cluster'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleByokCluster;
