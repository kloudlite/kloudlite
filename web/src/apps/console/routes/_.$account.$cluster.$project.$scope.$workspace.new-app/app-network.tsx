import { ArrowRight, X } from '@jengaicons/react';
import { useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { TextInput } from '~/components/atoms/input';
import List from '~/console/components/list';
import { FadeIn } from './util';

interface IExposedPorts {
  targetPort: string;
  port: string;
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
  const [port, setPort] = useState<string>('');
  const [targetPort, setTargetPort] = useState<string>('');
  const [exposedPorts, setExposedPorts] = useState<Array<IExposedPorts>>([]);
  const [portError, setPortError] = useState<string>('');

  return (
    <>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-default">
        <div className="flex flex-row gap-3xl items-center">
          <div className="flex-1">
            <TextInput
              label="Target port"
              size="lg"
              autoComplete="off"
              value={targetPort}
              onChange={({ target }) => {
                setTargetPort(target.value);
              }}
            />
          </div>
          <div className="flex-1">
            <TextInput
              label="Exposed port"
              size="lg"
              error={!!portError}
              message={portError}
              value={port}
              onChange={({ target }) => {
                setPort(target.value);
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
            onClick={() => {
              if (exposedPorts.find((ep) => ep.targetPort === targetPort)) {
                setPortError('Port is already exposed.');
              } else {
                setExposedPorts((prev) => [
                  ...prev,
                  { name: port, port, targetPort },
                ]);
                setPort('');
                setTargetPort('');
              }
            }}
          />
        </div>
      </div>
      {exposedPorts && exposedPorts.length > 0 && (
        <ExposedPortList
          exposedPorts={exposedPorts}
          onDelete={(ep) => {
            setExposedPorts(exposedPorts.filter((p) => p !== ep));
          }}
        />
      )}
    </>
  );
};

const AppNetwork = () => {
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
