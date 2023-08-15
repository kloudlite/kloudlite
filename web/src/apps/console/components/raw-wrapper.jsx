import Tooltip from '~/components/atoms/tooltip';

const RawWrapper = ({ leftChildren, rightChildren }) => {
  return (
    <Tooltip.Provider>
      <div className="min-h-full flex flex-row">
        <div className="min-h-full flex flex-col bg-surface-basic-subdued px-11xl pt-11xl pb-10xl">
          <div className="flex flex-col items-start gap-6xl w-[379px]">
            {leftChildren}
          </div>
        </div>
        <div className="pt-11xl pb-12xl px-11xl flex flex-1 bg-surface-basic-default">
          <div className="w-[549px]">{rightChildren}</div>
        </div>
      </div>
    </Tooltip.Provider>
  );
};

export default RawWrapper;
