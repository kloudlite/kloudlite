import { ArrowRight, X } from '@jengaicons/react';
import { useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { NumberInput, TextInput } from '~/components/atoms/input';
import List from '~/console/components/list';
import {
  FadeIn,
  InfoLabel,
  parseValue,
} from '../_.$account.$cluster.$project.$scope.$workspace.new-app/util';

interface IExposedPorts {
  targetPort?: number;
  port: number;
}

interface IExposedPortList {
  exposedPorts: IExposedPorts[];
  onDelete: (exposedPorts: IExposedPorts) => void;
}
const ExposedPortList = ({
  exposedPorts,
  onDelete = (_) => _,
}: IExposedPortList) => {
  return (
    <div className="flex flex-col gap-lg">
      <div className="text-text-strong bodyMd">Exposed port</div>
      <List.Root>
        {exposedPorts.map((ep, index) => {
          return (
            <List.Row
              key={ep.port}
              columns={[
                {
                  key: `${index}-column-2`,
                  className: 'flex-1',
                  render: () => (
                    <div className="flex flex-row gap-md items-center bodyMd text-text-soft">
                      <span>Service: </span>
                      {ep.port}
                      <ArrowRight size={16} weight={1} />
                      <span>Container: </span>
                      {ep.targetPort}
                    </div>
                  ),
                },
                {
                  key: `${index}-column-3`,
                  render: () => (
                    <div>
                      <IconButton
                        icon={<X />}
                        variant="plain"
                        size="sm"
                        onClick={() => {
                          onDelete(ep);
                        }}
                      />
                    </div>
                  ),
                },
              ]}
            />
          );
        })}
      </List.Root>
    </div>
  );
};

const ExposedPorts = () => {
  const [port, setPort] = useState<number>(3000);
  const [targetPort, setTargetPort] = useState<number>(3000);
  const [portError, setPortError] = useState<string>('');

  //   const { services, setServices } = useAppState();

  return (
    <>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-default">
        <TextInput label="Name" size="lg" />
        <div className="flex flex-row gap-3xl items-center">
          <div className="flex-1">
            <NumberInput
              label={
                <InfoLabel label="Expose Port" info="info about expose port" />
              }
              size="lg"
              error={!!portError}
              message={portError}
              value={port}
              onChange={({ target }) => {
                setPort(parseValue(target.value, 0));
              }}
            />
          </div>
          <div className="flex-1">
            <NumberInput
              min={0}
              max={65536}
              label={
                <InfoLabel
                  info="info about container port"
                  label="Container port"
                />
              }
              size="lg"
              autoComplete="off"
              value={targetPort}
              onChange={({ target }) => {
                setTargetPort(parseValue(target.value, 0));
              }}
            />
          </div>
        </div>
        <div className="flex flex-row gap-md items-center">
          <div className="bodySm text-text-soft flex-1">
            All network entries be mounted on the path specified in the
            container
          </div>
          <Button
            content="Expose port"
            variant="basic"
            disabled={!port || !targetPort}
          />
        </div>
      </div>
      {/* {services.length > 0 && (
        <ExposedPortList exposedPorts={[]} onDelete={(ep) => {}} />
      )} */}
    </>
  );
};

const AppNetwork = () => {
  //   const { setPage } = useAppState();
  return (
    <FadeIn>
      <div className="flex flex-col gap-xl ">
        <div className="headingXl text-text-default">Network</div>
        <div className="bodyMd text-text-soft">
          Expose service ports that need to be exposed from container
        </div>
      </div>
      <ExposedPorts />
    </FadeIn>
  );
};

export default AppNetwork;
