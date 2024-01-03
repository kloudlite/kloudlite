/* eslint-disable react/destructuring-assignment */
import * as Chips from '~/components/atoms/chips';
import { TextArea, TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import { yamlDump } from '~/console/components/diff-viewer';
import { IdSelector } from '~/console/components/id-selector';
import { IDialogBase } from '~/console/components/types.d';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IHelmCharts } from '~/console/server/gql/queries/helm-chart-queries';
import { ExtractNodeType, parseName } from '~/console/server/r-utils/common';
import { useReload } from '~/root/lib/client/helpers/reloader';
import useForm, { dummyEvent } from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { handleError } from '~/root/lib/utils/common';
import yaml from 'js-yaml';
import { useParams } from '@remix-run/react';
import axios from 'axios';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { useEffect, useState } from 'react';
import Select from '~/components/atoms/select';

type IDialog = IDialogBase<ExtractNodeType<IHelmCharts>>;

type IHelmDoc = {
  apiVersion: string;
  entries: {
    [key: string]: { version: string }[];
  };
  generated: string;
};
const Root = (props: IDialog) => {
  const { isUpdate, setVisible } = props;
  const api = useConsoleApi();
  const reloadPage = useReload();
  const { cluster } = useParams();
  const [repoNames, setRepoNames] = useState<
    Array<{ label: string; value: string }>
  >([]);
  const [repoNamesLoading, setRepoNamesLoading] = useState(false);

  const repoUrlChanges = async (repoUrl: string) => {
    try {
      setRepoNamesLoading(true);
      const res = await axios.get(`${repoUrl}/index.yaml`);
      const repos = yaml.load(res.data, { json: true }) as IHelmDoc;
      setRepoNames(
        Object.keys(repos.entries).map((v) => ({ label: v, value: v }))
      );
    } catch (error) {
      console.log(error);
    } finally {
      setRepoNamesLoading(false);
    }
  };

  const { values, errors, handleSubmit, handleChange, isLoading, resetValues } =
    useForm({
      initialValues: isUpdate
        ? {
            displayName: '',
            name: '',
            namespace: '',
            chartName: '',
            chartRepo: {
              name: '',
              url: '',
            },
            chartVersion: '',
            values: '',
          }
        : {
            displayName: '',
            name: '',
            namespace: '',
            chartName: '',
            chartRepo: {
              name: '',
              url: '',
            },
            chartVersion: '',
            values: '',
          },
      validationSchema: isUpdate
        ? Yup.object({
            displayName: Yup.string().required(),
            name: Yup.string().required(),
            namespace: Yup.string().required(),
            chartRepo: Yup.object({
              name: Yup.string(),
              url: Yup.string(),
            }),
          })
        : Yup.object({
            displayName: Yup.string().required(),
            name: Yup.string().required(),
            namespace: Yup.string().required(),
            chartRepo: Yup.object({
              name: Yup.string(),
              url: Yup.string(),
            }),
          }),

      onSubmit: async (val) => {
        console.log('yamlvalue', yaml.load(val.values, { json: true }));
        if (!cluster) {
          throw new Error('Cluster is required.');
        }
        try {
          const { errors } = await api.createHelmChart({
            clusterName: cluster || '',
            release: {
              displayName: val.displayName,
              metadata: {
                name: val.name,
              },
              spec: {
                chartName: val.chartName,
                chartVersion: val.chartVersion,
                chartRepo: val.chartRepo,
                values: yaml.load(val.values, { json: true }),
              },
            },
          });

          if (errors) {
            throw errors[0];
          }
          reloadPage();
          setVisible(false);
          resetValues();
        } catch (error) {
          handleError(error);
        }
      },
    });

  useDebounce(() => repoUrlChanges(values.chartRepo.url), 300, [
    values.chartRepo.url,
  ]);

  return (
    <Popup.Form onSubmit={handleSubmit}>
      <Popup.Content>
        <div className="flex flex-col gap-2xl">
          <TextInput
            label="Name"
            onChange={handleChange('displayName')}
            error={!!errors.displayName}
            message={errors.displayName}
            value={values.displayName}
            name="helm-name"
          />
          {isUpdate && (
            <Chips.Chip
              {...{
                item: { id: parseName(props.data) },
                label: parseName(props.data),
                prefix: 'Id:',
                disabled: true,
                type: 'BASIC',
              }}
            />
          )}
          {!isUpdate && (
            <IdSelector
              name={values.displayName}
              resType="cluster"
              onChange={(id) => {
                handleChange('name')({ target: { value: id } });
              }}
            />
          )}
          <TextInput
            label="Namespace"
            onChange={({ target }) => {
              handleChange('namespace')(dummyEvent(target.value));
            }}
            error={!!errors.namespace}
            message={errors.namespace}
            value={values.namespace}
            name="helm-chart-namespace"
          />
          <TextInput
            label="Chart repo url"
            onChange={({ target }) => {
              handleChange('chartRepo.url')(dummyEvent(target.value));
              setRepoNamesLoading(true);
            }}
            error={!!errors.chartRepo}
            message={errors.chartRepo}
            value={values.chartRepo.url}
            name="helm-chart-repo-url"
          />
          <Select
            label="Chart name"
            value={undefined}
            options={async () => repoNames}
            loading={repoNamesLoading}
          />
          <TextInput
            label="Chart version"
            onChange={handleChange('chartVersion')}
            error={!!errors.chartVersion}
            message={errors.chartVersion}
            value={values.chartVersion}
            name="helm-chart-version"
          />
          <TextArea
            label="Values"
            onChange={handleChange('values')}
            error={!!errors.values}
            message={errors.values}
            value={values.values}
            name="helm-chart-values"
          />
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button content="Cancel" variant="basic" closable />
        <Popup.Button
          loading={isLoading}
          type="submit"
          content={isUpdate ? 'Update' : 'Add'}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Form>
  );
};

const HandleHelmChart = (props: IDialog) => {
  const { isUpdate, setVisible, visible } = props;

  return (
    <Popup.Root show={visible} onOpenChange={(v) => setVisible(v)}>
      <Popup.Header>
        {isUpdate ? 'Edit helm chart' : 'Add new helm chart'}
      </Popup.Header>
      {(!isUpdate || (isUpdate && props.data)) && <Root {...props} />}
    </Popup.Root>
  );
};

export default HandleHelmChart;
