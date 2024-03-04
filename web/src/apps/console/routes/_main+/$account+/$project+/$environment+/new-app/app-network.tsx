import { useEffect, useState } from 'react';
import { Button, IconButton } from '~/components/atoms/button';
import { NumberInput } from '~/components/atoms/input';
import { usePagination } from '~/components/molecule/pagination';
import { cn } from '~/components/utils';
import List from '~/console/components/list';
import NoResultsFound from '~/console/components/no-results-found';
import { useAppState } from '~/console/page-components/app-states';
import { BottomNavigation, InfoLabel } from '~/console/components/commons';
import { useUnsavedChanges } from '~/root/lib/client/hooks/use-unsaved-changes';
import {
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  SmileySad,
  X,
} from '~/console/components/icons';
import { FadeIn, parseValue } from '~/console/page-components/util';

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
  const { page, hasNext, hasPrevious, onNext, onPrev, setItems } =
    usePagination({
      items: exposedPorts,
      itemsPerPage: 5,
    });

  useEffect(() => {
    setItems(exposedPorts);
  }, [exposedPorts]);
  return (
    <div className="flex flex-col gap-lg bg-surface-basic-default">
      {exposedPorts.length > 0 && (
        <List.Root
          className="min-h-[347px] !shadow-none"
          header={
            <div className="flex flex-row items-center justify-between w-full">
              <div className="text-text-strong bodyMd flex-1">
                Exposed ports
              </div>
              <div className="flex flex-row items-center">
                <IconButton
                  icon={<ChevronLeft />}
                  size="xs"
                  variant="plain"
                  onClick={() => onPrev()}
                  disabled={!hasPrevious}
                />
                <IconButton
                  icon={<ChevronRight />}
                  size="xs"
                  variant="plain"
                  onClick={() => onNext()}
                  disabled={!hasNext}
                />
              </div>
            </div>
          }
        >
          {page.map((ep, index) => {
            return (
              <List.Row
                className={cn({
                  '!border-b': index < 4,
                  '!rounded-b-none': index < 4,
                })}
                key={ep.port}
                columns={[
                  {
                    key: `${index}-column-2`,
                    className: 'flex-1',
                    render: () => (
                      <div className="flex flex-row gap-md items-center bodyMd text-text-soft">
                        <span>Service: </span>
                        {ep.port}
                        <ArrowRight size={16} />
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
      )}
      {exposedPorts.length === 0 && (
        <div className="rounded border-border-default border min-h-[347px] flex flex-row items-center justify-center">
          <NoResultsFound
            title={null}
            subtitle="No ports are exposed currently"
            compact
            image={<SmileySad size={32} weight={1} />}
            shadow={false}
            border={false}
          />
        </div>
      )}
    </div>
  );
};

export const ExposedPorts = () => {
  const [port, setPort] = useState<number | string>('');
  const [targetPort, setTargetPort] = useState<number | string>('');
  const [portError, setPortError] = useState<string>('');

  const { services, setServices } = useAppState();

  // for updating
  const { hasChanges } = useUnsavedChanges();

  // for updating
  useEffect(() => {
    if (!hasChanges) {
      setPort('');
      setTargetPort('');
      setPortError('');
    }
  }, [hasChanges]);

  return (
    <>
      <div className="flex flex-col gap-3xl p-3xl rounded border border-border-default">
        <div className="flex flex-row gap-3xl items-start">
          <div className="flex-1">
            <NumberInput
              min={0}
              max={65534}
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
              max={65534}
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
            onClick={() => {
              if (services?.find((ep) => ep.port && ep.port === port)) {
                setPortError('Port is already exposed.');
              } else {
                if (typeof port === 'number' && typeof targetPort === 'number')
                  setServices((prev) => [
                    ...prev,
                    {
                      name: `port-${port}`,
                      port,
                      targetPort,
                    },
                  ]);
                setPort('');
                setTargetPort('');
                setPortError('');
              }
            }}
          />
        </div>
      </div>
      <ExposedPortList
        exposedPorts={services}
        onDelete={(ep) => {
          setServices((s) => {
            return s.filter((v) => v.port !== ep.port);
          });
        }}
      />
    </>
  );
};

const AppNetwork = () => {
  const { setPage, markPageAsCompleted } = useAppState();
  return (
    <FadeIn notForm>
      <div className="bodyMd text-text-soft">
        Expose service ports that need to be exposed from container
      </div>

      <ExposedPorts />
      <BottomNavigation
        primaryButton={{
          type: 'submit',
          content: 'Save & Continue',
          variant: 'primary',
          onClick: () => {
            setPage(5);
            markPageAsCompleted(4);
            markPageAsCompleted(5);
          },
        }}
        secondaryButton={{
          content: 'Environments',
          variant: 'outline',
          onClick: () => {
            setPage(3);
          },
        }}
      />
    </FadeIn>
  );
};

export default AppNetwork;
