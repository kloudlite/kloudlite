import { useAppState } from '~/console/page-components/app-states';
import { parseValue } from '~/console/page-components/util';
import { keyconstants } from '~/console/server/r-utils/key-constants';
import { AppIn } from '~/root/src/generated/gql/server';

export const appInitialFormValues = ({
  app,
  getContainer,
}: {
  app: ReturnType<typeof useAppState>['app'];
  getContainer: ReturnType<typeof useAppState>['getContainer'];
}) => {
  return {
    imagePullPolicy: getContainer().imagePullPolicy || 'IfNotPresent',
    pullSecret: 'TODO',
    cpuMode: app.metadata?.annotations?.[keyconstants.cpuMode] || 'shared',
    memPerCpu: app.metadata?.annotations?.[keyconstants.memPerCpu] || '1',

    autoscaling: app.spec.hpa?.enabled || false,
    minReplicas: app.spec.hpa?.minReplicas || 1,
    maxReplicas: app.spec.hpa?.maxReplicas || 3,
    cpuThreshold: app.spec.hpa?.thresholdCpu || 75,
    memoryThreshold: app.spec.hpa?.thresholdMemory || 75,
    replicas: app.spec.replicas || 1,

    cpu: parseValue(getContainer().resourceCpu?.max, 250),

    selectedPlan:
      app.metadata?.annotations[keyconstants.selectedPlan] || 'shared-1',
    selectionMode:
      app.metadata?.annotations[keyconstants.selectionModeKey] || 'quick',
    manualCpuMin: parseValue(getContainer().resourceCpu?.min, 0),
    manualCpuMax: parseValue(getContainer().resourceCpu?.max, 0),
    manualMemMin: parseValue(getContainer().resourceMemory?.min, 0),
    manualMemMax: parseValue(getContainer().resourceMemory?.max, 0),

    nodepoolName: app.spec.nodeSelector?.[keyconstants.nodepoolName] || '',
  };
};

export const mapFormValuesToApp = ({
  oldAppIn: s,
  appIn: val,
}: {
  oldAppIn: AppIn;
  appIn: ReturnType<typeof appInitialFormValues>;
}): AppIn => {
  return {
    ...s,
    metadata: {
      ...s.metadata!,
      annotations: {
        ...(s.metadata?.annotations || {}),
        [keyconstants.cpuMode]: val.cpuMode,
        [keyconstants.selectedPlan]: val.selectedPlan,
        [keyconstants.selectionModeKey]: val.selectionMode,
        [keyconstants.memPerCpu]: `${val.memPerCpu}`,
      },
    },
    spec: {
      ...s.spec,
      nodeSelector: val.nodepoolName
        ? {
            ...(s.spec.nodeSelector || {}),
            [keyconstants.nodepoolName]: val.nodepoolName,
          }
        : null,

      containers: [
        {
          ...(s.spec.containers?.[0] || {}),
          imagePullPolicy: val.imagePullPolicy,
          resourceCpu:
            val.selectionMode === 'quick'
              ? {
                  max: `${val.cpu}m`,
                  min:
                    val.cpuMode === 'shared'
                      ? `${val.cpu / 4}m`
                      : `${val.cpu}m`,
                }
              : {
                  max: `${val.manualCpuMax}m`,
                  min: `${val.manualCpuMin}m`,
                },
          resourceMemory:
            val.selectionMode === 'quick'
              ? {
                  max: `${(
                    parseValue(val.cpu, 1) * parseValue(val.memPerCpu, 4)
                  ).toFixed(0)}Mi`,
                  min: `${val.cpu}Mi`,
                }
              : {
                  max: `${val.manualMemMax}Mi`,
                  min: `${val.manualMemMin}Mi`,
                },
        },
      ],

      hpa: {
        enabled: val.autoscaling,
        maxReplicas: val.maxReplicas,
        minReplicas: val.minReplicas,
        thresholdCpu: val.cpuThreshold,
        thresholdMemory: val.memoryThreshold,
      },
      replicas: val.replicas,
    },
  };
};

export default appInitialFormValues;
