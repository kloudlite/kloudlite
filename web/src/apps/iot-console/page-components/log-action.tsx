import { cn } from '@kloudlite/design-system/utils';
import { Clock, ListNumbers } from '~/iotconsole/components/icons';
import { useDataState } from './common-state';

const LogAction = () => {
  const { state, setState } = useDataState<{
    linesVisible: boolean;
    timestampVisible: boolean;
  }>('logs');

  return (
    <div className="hljs flex items-center gap-xl px-xs">
      <div
        onClick={() => {
          setState((s) => ({ ...s, linesVisible: !s.linesVisible }));
        }}
        className="flex items-center justify-center font-bold text-xl cursor-pointer select-none active:translate-y-[1px] transition-all"
      >
        <span
          className={cn({
            'opacity-50': !state.linesVisible,
          })}
        >
          <ListNumbers color="currentColor" size={16} />
        </span>
      </div>
      <div
        onClick={() => {
          setState((s) => ({
            ...s,
            timestampVisible: !s.timestampVisible,
          }));
        }}
        className="flex items-center justify-center font-bold text-xl cursor-pointer select-none active:translate-y-[1px] transition-all"
      >
        <span
          className={cn({
            'opacity-50': !state.timestampVisible,
          })}
        >
          <Clock color="currentColor" size={16} />
        </span>
      </div>
    </div>
  );
};

export default LogAction;
