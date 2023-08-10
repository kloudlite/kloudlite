import { Button as NativeButton } from '~/components/atoms/button';

export const Breadcrum = ({ children }) => {
  return <div className="flex flex-row gap-md items-center">{children}</div>;
};

export const Button = (props) => {
  return (
    <div className="flex flex-row gap-md items-center">
      <div className="text-text-disabled bodySm">/</div>
      <NativeButton {...props} size="md" variant="plain" />
    </div>
  );
};

export default Breadcrum;
