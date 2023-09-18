import { ReactNode } from 'react';

interface IMultiStepForm {
  items: Array<ReactNode>;
  onNext?: () => void;
  onPrevious?: () => void;
}
const MultiStepForm = ({ items = [], onNext, onPrevious }: IMultiStepForm) => {
  return <div>Multisrtep</div>;
};

export default MultiStepForm;
