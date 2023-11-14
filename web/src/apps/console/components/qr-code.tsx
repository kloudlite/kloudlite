import QRCode from 'react-qr-code';

interface IQRCodeView {
  value: string;
}
const QRCodeView = ({ value }: IQRCodeView) => {
  // @ts-ignore
  // eslint-disable-next-line react/jsx-pascal-case
  return <QRCode.default value={value} />;
};

export default QRCodeView;
