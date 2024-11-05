import { ReactNode, useState } from 'react';

interface IUseSteps {
  items: Array<ReactNode>;
  defaultItem?: ReactNode;
}
const useSteps = ({ items = [], defaultItem }: IUseSteps) => {
  const [currentItem, setCurrentItem] = useState(defaultItem || items[0]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const onNext = () => {
    if (currentIndex + 1 <= items.length) {
      setCurrentItem(items[currentIndex + 1]);
      setCurrentIndex((prev) => prev + 1);
    }
  };

  const onPrevious = () => {
    if (currentIndex - 1 >= 0) {
      setCurrentItem(items[currentIndex - 1]);
      setCurrentIndex((prev) => prev - 1);
    }
  };
  return { item: currentItem, onNext, onPrevious };
};

export default useSteps;
