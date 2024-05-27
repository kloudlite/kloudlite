import { NN } from '~/root/lib/types/common';
import { ExtractArrayType } from '~/console/page-components/util';
import { IExternalApp } from '~/console/server/gql/queries/external-app-queries';
import KeyValuePair from '~/console/components/key-value-pair';
import { dummyEvent } from '~/root/lib/client/hooks/use-form';

export type exposedExternalAppPortsType = ExtractArrayType<
  NN<NN<NN<IExternalApp['spec']>['intercept']>['portMappings']>
>;

interface IExposedExternalAppPortList {
  handleChange: (key: string) => (e: {
    target: {
      value: string;
    };
  }) => void;
  values: Record<string, any>;
}
const ExposedExternalAppPortList = ({
  handleChange,
  values,
}: IExposedExternalAppPortList) => {
  return (
    <div className="flex flex-col gap-lg bg-surface-basic-default">
      <KeyValuePair
        type="number"
        addText="Add new port"
        label="Exposed Ports"
        size="lg"
        keyLabel="appPort"
        valueLabel="devicePort"
        keyPlaceholder="App Port"
        valuePlaceholder="Device Port"
        value={values.appPortsTemp}
        onChange={(val, __, v) => {
          handleChange('appPorts')(dummyEvent(v));
          handleChange('appPortsTemp')(dummyEvent(val));
        }}
      />
    </div>
  );
};

export default ExposedExternalAppPortList;
