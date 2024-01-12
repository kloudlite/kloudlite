import {
  Dispatch,
  ReactNode,
  SetStateAction,
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';

type IStep = {
  label: string;
  children: ReactNode;
};

interface IModifiedStep extends IStep {
  active: boolean;
  completed: boolean;
}

const ManagedServiceStateContext = createContext<{
  nextStep: () => void;
  prevStep: () => void;
  steps: IModifiedStep[];
  setSteps: Dispatch<SetStateAction<IStep[]>>;
}>({
  nextStep() {},
  prevStep() {},
  steps: [],
  setSteps() {},
});

export const ManagedServiceStateProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  const [currentStepNumber, setCurrentStepNumber] = useState(1);
  const [orgSteps, setOrgSteps] = useState<IStep[]>([]);
  const [steps, setModifiedSteps] = useState<IModifiedStep[]>([]);

  const nextStep = () => {
    if (currentStepNumber < orgSteps.length - 1) {
      setCurrentStepNumber((prev) => prev + 1);
    }
  };

  const prevStep = () => {
    if (currentStepNumber > 0) {
      setCurrentStepNumber((prev) => prev - 1);
    }
  };

  useEffect(() => {
    setModifiedSteps([
      ...orgSteps.map((step, index) => ({
        ...step,
        children: index + 1 === currentStepNumber ? step.children : null,
        active: index + 1 === currentStepNumber,
        completed: false,
      })),
    ]);
  }, [currentStepNumber, orgSteps]);

  return (
    <ManagedServiceStateContext.Provider
      value={useMemo(
        () => ({
          nextStep,
          prevStep,
          steps,
          setSteps: setOrgSteps,
        }),
        [steps]
      )}
    >
      {children}
    </ManagedServiceStateContext.Provider>
  );
};
const useManagedServiceState = ({ steps = [] }: { steps: IStep[] }) => {
  const context = useContext(ManagedServiceStateContext);
  console.log(steps);

  useEffect(() => {
    context.setSteps(steps);
  }, []);
  return context;
};
export default useManagedServiceState;
